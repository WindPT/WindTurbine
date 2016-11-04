package main

import (
    "bytes"
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "math"
    "net"
    "os"
    "regexp"
    "strings"
    "time"
    "github.com/jackpal/bencode-go"
    "github.com/jinzhu/gorm"
    "github.com/kataras/iris"
    "github.com/kinosang/php_serialize"
    "github.com/Knetic/govaluate"
    "github.com/oleiade/reflections"
    _ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
    xmlFile, err := os.Open("config.xml")

    if err != nil {
        fmt.Println(time.Now().Format(time.RFC3339), "Faild to read config file:", err)
        return
    }

    defer xmlFile.Close()

    b, _ := ioutil.ReadAll(xmlFile)

    var s Setting
    xml.Unmarshal(b, &s)

    gorm.DefaultTableNameHandler = func (db *gorm.DB, defaultTableName string) string  {
        switch defaultTableName {
            case "common_configs": defaultTableName = "common_config"
            case "user_bans": defaultTableName = "user_ban"
            case "user_datas": defaultTableName = "user_data"
            case "users": defaultTableName = "user"
            case "windid_user_datas": defaultTableName = "windid_user_data"
        }
        return s.TablePrefix + defaultTableName;
    }

    db, err := gorm.Open("mysql", s.DSN)
    db.LogMode(s.Debug)

    if err != nil {
        panic(err)
    }

    defer db.Close()

    var user_agents []AppTorrentAgent
    db.Order("id").Find(&user_agents)

    var common_config CommonConfig
    db.Where("name = \"app.torrent.credits\"").First(&common_config)

    decoder := php_serialize.NewUnSerializer(common_config.Value)
    exp_pvalue, err := decoder.Decode()

    if err != nil {
        panic(err)
    }

    exp_array, _ := exp_pvalue.(php_serialize.PhpArray)
    credits := make(map[int]Credit)

    for k, v := range exp_array {
        v_array := v.(php_serialize.PhpArray)
        v_enabled := v_array["enabled"].(string)
        v_exp := v_array["exp"].(string)

        credits[k.(int)] = Credit {enabled: v_enabled == "1", exp: v_exp}
    }

    tr := &TrackerResource {db: db, user_agents: user_agents, credits: credits}

    iris.OnError(iris.StatusNotFound, func(c *iris.Context) {
        berror(c, "错误：Passkey 不能为空")
    })

    iris.Get("/:passkey", tr.Announcement)

    iris.Listen(s.Listen)
}

func (tr *TrackerResource) Announcement(c *iris.Context) {
    user_agent := string(c.UserAgent())
    passkey := c.Param("passkey")

    // Check parameters
    m := c.URLParams()
    required_params := []string{"info_hash","peer_id","port","uploaded","downloaded","left"}
    for _,param_key := range required_params {
        if _, ok := m[param_key]; !ok {
            berror(c, fmt.Sprintf("错误：缺少参数 %s", param_key))
            return
        }
    }

    event         := c.URLParam("event")
    info_hash     := c.URLParam("info_hash")
    peer_id       := c.URLParam("peer_id")
    port, _       := c.URLParamInt("port")
    uploaded, _   := c.URLParamInt("uploaded")
    downloaded, _ := c.URLParamInt("downloaded")
    left, _       := c.URLParamInt("left")

    state := "started"

    // Get client IP
    ips := strings.Split(c.RequestHeader("X-FORWARDED-FOR"), ", ")
    ip := ips[0]
    if ip == "" {
       ip = c.RequestIP()
    }

    if ip == "" {
        berror(c, "错误：无法获取客户端IP")
        return
    }

    // Check UserAgent
    allowed := false
    for _, v := range tr.user_agents {
        if allowed, _ = regexp.MatchString(v.AgentPattern, user_agent); allowed {
            if len(v.PeerIdPattern) > 0 {
                if allowed, _ = regexp.MatchString(v.PeerIdPattern, peer_id); allowed {
                    break
                }
            } else {
                break
            }
        }
    }

    if !allowed {
        berror(c, "错误：客户端不被支持")
        return
    }

    // Get user info by passkey
    var user AppTorrentUser
    tr.db.Where("passkey = ?", passkey).First(&user)

    if user.Passkey != passkey {
        berror(c, "错误：无效的 passkey，请尝试重新下载种子")
        return
    }

    var pwuser User
    tr.db.Where("uid = ?", user.Uid).First(&pwuser)

    if user.Uid != pwuser.Uid {
        berror(c, "错误：用户不存在，请尝试重新下载种子")
        return
    }

    // Check if user is banned
    var user_ban UserBan
    tr.db.Where("uid = ?", user.Uid).First(&user_ban)

    if user_ban.Uid == user.Uid {
        berror(c, fmt.Sprintf("错误：用户已被封禁 %s", user_ban.Reason))
        return
    }

    // Get torrent info by info_hash
    var torrent AppTorrent
    tr.db.Where("info_hash = ?", info_hash).First(&torrent)

    if torrent.InfoHash != info_hash {
        berror(c, "错误：种子信息未注册，可能是已被删除")
        return
    }

    var bbs_thread BbsThread
    tr.db.Where("tid = ?", torrent.Tid).First(&bbs_thread)

    if bbs_thread.Tid != torrent.Tid {
        berror(c, "错误：种子不存在")
        return
    }

    if bbs_thread.Disabled > 0 && bbs_thread.CreatedUserid != user.Uid {
        if pwuser.Groupid < 3 || pwuser.Groupid > 5 {
            berror(c, "错误：种子已删除或待审核")
            return
        }
    }

    // Get peers list by torrent
    seeders := 0
    leechers := 0
    var self AppTorrentPeer
    var peers []AppTorrentPeer
    tr.db.Where("torrent_id = ?", torrent.Id).Find(&peers)

    i := 0
    for _, peer := range peers {
        if peer.Seeder {
            seeders++
        } else {
            leechers++
        }

        if peer.PeerId == peer_id {
            self = peer
            peers = append(peers[:i], peers[i+1:]...)
        }

        i++
    }

    // Update peer info
    if (AppTorrentPeer{}) == self {
        self.TorrentId = torrent.Id
        self.Uid = user.Uid
        self.Ip = ip
        self.PeerId = peer_id
        self.Port = port
        self.Uploaded = uploaded
        self.Downloaded = downloaded
        self.Left = left
        self.Agent = user_agent
        self.StartedAt = time.Now()
        self.LastAction = time.Now()
    }

    self.Seeder = left <= 0

    if self.Seeder {
        seeders++
    } else {
        leechers++
    }

    conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
    if err != nil {
        self.Connectable = false
    } else {
        self.Connectable = true
        defer conn.Close()
    }

    // Check if already started
    if self.PeerId != peer_id || self.Ip != ip {
        berror(c, "错误：同一种子禁止多处下载")
        return
    }

    switch event {
        case "", "started": {
            self.Port = port
            self.Uploaded = uploaded
            self.Downloaded = downloaded
            self.Left = left
            self.Agent = user_agent
            self.LastAction = time.Now()

            tr.db.Save(&self)
        }
        case "stopped": {
            if (self.Id != 0) {
                tr.db.Delete(&self)
            }

            state = "stopped"
        }
        case "completed": {
            self.Port = port
            self.Uploaded = uploaded
            self.Downloaded = downloaded
            self.Left = left
            self.Agent = user_agent
            self.FinishedAt = time.Now()
            self.LastAction = time.Now()

            tr.db.Save(&self)
        }
        default: {
            berror(c, "错误：客户端发送未知状态")
            return
        }
    }

    // Update history
    var rotio float64
    uploaded_add := 0
    downloaded_add := 0
    uploaded_total := uploaded
    downloaded_total := downloaded

    var history AppTorrentHistory
    tr.db.Where("torrent_id = ? AND uid = ?", torrent.Id, user.Uid).First(&history)

    if history.Uid != user.Uid || history.TorrentId != torrent.Id {
        history = AppTorrentHistory{
            Uid: user.Uid,
            TorrentId: torrent.Id,
            Uploaded: uploaded,
            UploadedLast: uploaded,
            Downloaded: downloaded,
            DownloadedLast: downloaded,
            Left: left,
            State: state,
        }

        tr.db.Create(&history)
    } else {
        uploaded_add = int(math.Max(0, float64(uploaded - history.UploadedLast)))
        downloaded_add = int(math.Max(0, float64(downloaded - history.DownloadedLast)))

        uploaded_total = history.Uploaded + uploaded_add
        downloaded_total = history.Downloaded + downloaded_add

        history.Uploaded = uploaded_total
        history.UploadedLast = uploaded
        history.Downloaded = downloaded_total
        history.DownloadedLast = downloaded
        history.Left = left
        history.State = state

        if state == "stopped" {
            history.UploadedLast = 0
            history.DownloadedLast = 0
        }

        tr.db.Save(&history)
    }

    if downloaded_total != 0 {
        rotio = math.Floor(float64(uploaded_total / downloaded_total * 100) + 0.5) / 100
    } else {
        rotio = 1
    }

    // Update credits
    parameters := make(map[string]interface{}, 17)

    parameters["e"] = math.E
    parameters["pi"] = math.Pi
    parameters["phi"] = math.Phi

    var seeding []AppTorrentPeer
    tr.db.Where("uid = ? AND seeder = 1", user.Uid).Find(&seeding)

    var leeching []AppTorrentPeer
    tr.db.Where("uid = ? AND seeder = 0", user.Uid).Find(&leeching)

    var published_torrents []AppTorrent
    tr.db.Where("Owner = ?", user.Uid).Find(&published_torrents)

    var user_data UserData
    tr.db.Where("uid = ?", user.Uid).Find(&user_data)

    var windid_user_data WindidUserData
    tr.db.Where("uid = ?", user.Uid).Find(&windid_user_data)

    parameters["alive"] = time.Since(torrent.CreatedAt).Hours() / 24
    parameters["seeders"] = seeders
    parameters["leechers"] = leechers
    parameters["size"] = torrent.Size
    parameters["seeding"] = len(seeding)
    parameters["leeching"] = len(leeching)
    parameters["downloaded"] = uploaded_total
    parameters["downloaded_add"] = downloaded_add
    parameters["uploaded"] = uploaded_total
    parameters["uploaded_add"] = uploaded_add
    parameters["rotio"] = rotio
    parameters["time"] = time.Since(self.StartedAt).Seconds()
    parameters["time_la"] = time.Since(self.LastAction).Seconds()
    parameters["torrents"] = len(published_torrents)

    for k, v := range tr.credits {
        if !v.enabled {
            continue
        }

        credit_key := fmt.Sprintf("Credit%d", k)
        parameters["credit"], _ = reflections.GetField(user_data, credit_key)

        expression, _ := govaluate.NewEvaluableExpressionWithFunctions(v.exp, functions)

        delta, _ := expression.Evaluate(parameters)
        result := parameters["credit"].(float64) + delta.(float64)

        reflections.SetField(&user_data, credit_key, result)
        reflections.SetField(&windid_user_data, credit_key, result)
    }

    tr.db.Save(&user_data)
    tr.db.Save(&windid_user_data)

    // Update torrent peers count
    torrent.Seeders = seeders
    torrent.Leechers = leechers
    torrent.UpdatedAt = time.Now()

    tr.db.Save(&torrent)

    // Output peers list to client
    peer_list := PeerList{
        Interval: 840,
        MinInterval: 30,
        Complete: seeders,
        Incomplete: leechers,
        Peers: peers,
    }

    buf := new(bytes.Buffer)

    bencode.Marshal(buf, peer_list)

    c.Text(200, buf.String())
}

func berror(c *iris.Context, msg string){
    err := Error{msg}

    buf := new(bytes.Buffer)

    bencode.Marshal(buf, err)

    c.Text(200, buf.String())
}

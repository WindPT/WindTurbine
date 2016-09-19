package main

import (
    _ "github.com/go-sql-driver/mysql"
    "github.com/jackpal/bencode-go"
    "database/sql"
    "encoding/xml"
    "fmt"
    "io/ioutil"
    "math"
    "net"
    "net/http"
    "os"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type Error struct {
    reason string "failure reason";
}

type Setting struct {
    DSN string
    TablePrefix string
    Listen string
}

type UserAgent struct {
    family string
    peer_id_pattern string
    agent_pattern string
    allowhttps string
}

type User struct {
    uid int
    groupid int
    passkey string
    uploaded_mo int
    downloaded_mo int
}

type Torrent struct {
    id int
    tid int
    info_hash string
    leechers int
    seeders int
    size int
    added time.Time
}

type Peerinfo struct {
    id int
    uid int
    peer_id string
    ip string
    port int
    seeder string
    started time.Time
    last_action time.Time
}

type History struct {
    id int
    uid int
    torrent int
    uploaded int
    uploaded_last int
    downloaded int
    downloaded_last int
    status string
}

type Peer struct {
    ip string
    peer_id string "peer id"
    port int
}

type PeerList struct {
    interval int
    min_interval int "min interval"
    complete int
    incomplete int
    peers []Peer
}

func error(w http.ResponseWriter, msg string) {
    var err = Error{msg}
    bencode.Marshal(w, err)
}

func handler(db *sql.DB, allowedClients []UserAgent) func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {

        // Get passkey
        passkey := r.URL.Path[1:]
        if passkey == "" {
            error(w, "错误：Passkey 不能为空")
            return
        }

        // Check parameters
        m := r.URL.Query()
        required_params := []string{"info_hash","peer_id","port","uploaded","downloaded","left"}
        for _,param_key := range required_params {
            if _, ok := m[param_key]; !ok {
                error(w, fmt.Sprintf("错误：缺少参数 %s", param_key))
                return
            }
        }

        event         := m.Get("event")
        info_hash     := m.Get("info_hash")
        peer_id       := m.Get("peer_id")
        port, _       := strconv.Atoi(m.Get("port"))
        uploaded, _   := strconv.Atoi(m.Get("uploaded"))
        downloaded, _ := strconv.Atoi(m.Get("downloaded"))
        left, _       := strconv.Atoi(m.Get("left"))

        var seeder string
        if left > 0 {
            seeder = "no"
        } else {
            seeder = "yes"
        }

        var status string
        if left > 0 {
            status = "do"
        } else {
            status = "done"
        }

        // Get client IP
        ips := strings.Split(r.Header.Get("X-FORWARDED-FOR"), ", ")
        ip := ips[0]
        if ip == "" {
           ip = r.RemoteAddr
           ip, _, _ = net.SplitHostPort(ip)
        }

        if ip == "" {
            error(w, "错误：无法获取客户端IP")
            return
        }

        // Check UserAgent
        allowed := false
        var clientFamily string

        for _, v := range allowedClients {
            if allowed, _ = regexp.MatchString(v.agent_pattern, r.UserAgent()); allowed {
                if v.peer_id_pattern != "" {
                    allowed, _ = regexp.MatchString(v.peer_id_pattern, peer_id)
                }
            }

            if allowed {
                clientFamily = v.family
                break
            }
        }

        if !allowed {
            error(w, "错误：客户端不被支持")
            return
        }

        // Get user info by passkey
        var user User
        rows, _ := db.Query("SELECT `pw_user`.`uid` as `uid`, `groupid`, `passkey`, `uploaded_mo`, `downloaded_mo` FROM pw_app_torrent_user INNER JOIN pw_user ON pw_user.uid = pw_app_torrent_user.uid WHERE `passkey` = ?", passkey)
        rows.Next()
        rows.Scan(&user.uid, &user.groupid, &user.passkey, &user.uploaded_mo, &user.downloaded_mo)

        if user.passkey != passkey {
            error(w, "错误：无效的 passkey，请尝试重新下载种子")
            return
        }

        // Check if user is banned
        var uid int
        var reason string
        rows, _ = db.Query("SELECT `uid`, `reason` FROM pw_user_ban WHERE `uid` = ?", user.uid)
        rows.Next()
        rows.Scan(&uid, &reason)

        if uid == user.uid {
            error(w, fmt.Sprintf("错误：用户已被封禁 %s", reason))
            return
        }

        // Get torrent info by info_hash
        var torrent Torrent
        rows, _ = db.Query("SELECT `id`, `tid`, `info_hash`, `leechers`, `seeders`, `size`, `added` FROM pw_app_torrent WHERE `info_hash` = ?", info_hash)
        rows.Next()
        rows.Scan(&torrent.id, &torrent.tid, &torrent.info_hash, &torrent.leechers, &torrent.seeders, &torrent.size, &torrent.added)

        if torrent.info_hash != info_hash {
            error(w, "错误：种子信息未注册，可能是已被删除")
            return
        }

        // Check if torrent is disabled
        var tid int
        var disabled int
        var created_userid int
        rows, _ = db.Query("SELECT `tid`, `disabled`, `created_userid` FROM pw_bbs_threads WHERE `tid` = ?", torrent.tid)
        rows.Next()
        rows.Scan(&tid, &disabled, &created_userid)

        if tid != torrent.tid {
            error(w, "错误：种子已删除或待审核")
            return
        }

        if disabled > 0 && created_userid != user.uid {
            if user.groupid < 3 || user.groupid > 5 {
                error(w, "错误：种子已删除或待审核")
                return
            }
        }

        fmt.Println(time.Now().Format(time.RFC3339), "User", user.uid, "scrape for torrent", torrent.tid, "with", clientFamily)

        // Get peers list by torrent
        rows, _ = db.Query("SELECT `id`, `uid`, `peer_id`, `ip`, `port`, `seeder`, `started`, `last_action` FROM pw_app_torrent_peer WHERE `torrent` = ?", torrent.id)

        peers := []Peerinfo{}
        var self Peerinfo
        seeders := 0
        leechers := 0

        for rows.Next() {
            var r Peerinfo
            err := rows.Scan(&r.id, &r.uid, &r.peer_id, &r.ip, &r.port, &r.seeder, &r.started, &r.last_action)

            if err != nil {
                continue
            }

            if r.seeder == "yes" {
                seeders++
            } else {
                leechers++
            }

            // Get peer of current user specially
            if r.uid == user.uid {
                self = r
            } else {
                peers = append(peers, r)
            }
        }

        // Update peer info
        if self.peer_id != "" {
            // Check if already started
            if self.peer_id != peer_id || self.ip != ip {
                error(w, "错误：同一种子禁止多处下载")
                return
            }

            switch event {
                case "", "started": {
                    stmt, _ := db.Prepare("UPDATE pw_app_torrent_peer SET `port` = ?, `uploaded` = ?, `downloaded` = ?, `left` = ?, `last_action` = ?, `seeder` = ?, `agent` = ? WHERE `peer_id` = ?")
                    stmt.Exec(port, uploaded, downloaded, left, time.Now(), seeder, r.UserAgent(), self.peer_id)
                }
                case "stopped": {
                    stmt, _ := db.Prepare("DELETE FROM pw_app_torrent_peer WHERE `peer_id` = ?")
                    stmt.Exec(self.peer_id)
                    status = "stop"
                }
                case "completed": {
                    stmt, _ := db.Prepare("UPDATE pw_app_torrent_peer SET `finished_at` = ?, `port` = ?, `uploaded` = ?, `downloaded` = ?, `left` = ?, `last_action` = ?, `seeder` = ?, `agent` = ? WHERE `peer_id` = ?")
                    stmt.Exec(time.Now(), port, uploaded, downloaded, left, time.Now(), seeder, r.UserAgent(), self.peer_id)
                    status = "done"
                }
                default: {
                    error(w, "错误：客户端发送未知状态")
                    return
                }
            }
        } else {
            var connectable string
            _, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
            if err != nil {
                connectable = "no"
            } else {
                connectable = "yes"
            }

            stmt, _ := db.Prepare("INSERT INTO pw_app_torrent_peer (`torrent`, `uid`, `peer_id`, `ip`, `port`, `connectable`, `uploaded`, `downloaded`, `left`, `started`, `last_action`, `seeder`, `agent`) VALUE (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
            stmt.Exec(torrent.id, user.uid, peer_id, ip, port, connectable, uploaded, downloaded, left, time.Now(), time.Now(), seeder, r.UserAgent())
        }

        // Update user's history of this torrent
        var rotio float64
        uploaded_add := 0
        downloaded_add := 0
        uploaded_total := uploaded
        downloaded_total := downloaded

        var history History
        rows, _ = db.Query("SELECT * FROM pw_app_torrent_history WHERE `torrent` = ? AND `uid` = ?", torrent.id, user.uid)
        rows.Next()
        rows.Scan(&history.id, &history.uid, &history.torrent, &history.uploaded, &history.uploaded_last, &history.downloaded, &history.downloaded_last, &history.status)

        if history.uid != user.uid || history.torrent != torrent.id {
            stmt, _ := db.Prepare("INSERT INTO pw_app_torrent_history (`uid`, `torrent`, `uploaded`, `downloaded`) VALUE (?, ?, ?, ?)")
            stmt.Exec(user.uid, torrent.id, uploaded, downloaded)
        } else {
            uploaded_add = int(math.Max(0, float64(uploaded - history.uploaded_last)))
            downloaded_add = int(math.Max(0, float64(downloaded - history.downloaded_last)))

            uploaded_total = history.uploaded + uploaded_add
            downloaded_total = history.downloaded + downloaded_add

            stmt, _ := db.Prepare("UPDATE pw_app_torrent_history SET `uploaded` = ?, `uploaded_last` = ?, `downloaded` = ?, `downloaded_last` = ?, `status` = ? WHERE `id` = ?")
            stmt.Exec(uploaded_total, uploaded, downloaded_total, downloaded, status, history.id)
        }

        if downloaded_total != 0 {
            rotio = math.Floor(float64(uploaded_total / downloaded_total * 100) + 0.5) / 100
        } else {
            rotio = 1
        }

        // Update torrent peers count
        stmt, _ := db.Prepare("UPDATE pw_app_torrent SET `seeders` = ?, `leechers` = ?, `last_action` = ? WHERE `id` = ?")
        stmt.Exec(seeders, leechers, time.Now(), torrent.id)

        // Output peers list to client
        var peer_list PeerList
        peer_list.interval = 840
        peer_list.min_interval = 30
        peer_list.complete = seeders
        peer_list.incomplete = leechers

        for _, peer := range peers {
            var p Peer

            p.ip = peer.ip
            p.peer_id = peer.peer_id
            p.port = peer.port

            peer_list.peers = append(peer_list.peers, p)
        }

        fmt.Println(rotio)

        bencode.Marshal(w, peer_list)
    }
}

func main() {
    xmlFile, err := os.Open("config.xml")

    if err != nil {
        fmt.Println(time.Now().Format(time.RFC3339), "Error opening file:", err)
        return
    }

    defer xmlFile.Close()

    b, _ := ioutil.ReadAll(xmlFile)

    var s Setting
    xml.Unmarshal(b, &s)

    db, err := sql.Open("mysql", s.DSN)

    if err != nil {
        panic(err)
    }

    fmt.Println(time.Now().Format(time.RFC3339), "DB connection started")

    rows, _ := db.Query("SELECT `family`, `peer_id_pattern`, `agent_pattern`, `allowhttps` FROM pw_app_torrent_agent_allowed_family")

    if err != nil {
        panic(err)
    }

    allowedClients := []UserAgent{}

    for rows.Next() {
        var r UserAgent
        err := rows.Scan(&r.family, &r.peer_id_pattern, &r.agent_pattern, &r.allowhttps)

        if err != nil {
            panic(err)
        }

        allowedClients = append(allowedClients, r)
    }

    fmt.Println(time.Now().Format(time.RFC3339), "Configurations Loaded")

    fmt.Println(time.Now().Format(time.RFC3339), "Tracker starting at", s.Listen)

    http.HandleFunc("/", handler(db, allowedClients))
    http.ListenAndServe(s.Listen, nil)
}

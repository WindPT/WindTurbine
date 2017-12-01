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

	"github.com/Knetic/govaluate"
	"github.com/jackpal/bencode-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/kataras/iris"
	"github.com/kinosang/php_serialize"
	"github.com/oleiade/reflections"
)

func main() {
	// Read config file
	xmlFile, err := os.Open("config.xml")

	if err != nil {
		panic(err)
	}

	defer xmlFile.Close()

	b, _ := ioutil.ReadAll(xmlFile)

	var s Setting
	xml.Unmarshal(b, &s)

	// Initialize GORM
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		switch defaultTableName {
		case "common_configs":
			defaultTableName = "common_config"
		case "user_bans":
			defaultTableName = "user_ban"
		case "user_datas":
			defaultTableName = "user_data"
		case "users":
			defaultTableName = "user"
		case "windid_user_datas":
			defaultTableName = "windid_user_data"
		}
		return s.TablePrefix + defaultTableName
	}

	db, err := gorm.Open("mysql", s.DSN)
	db.LogMode(s.Debug)

	if err != nil {
		panic(err)
	}

	defer db.Close()

	// Load User-Agent whitelist
	var userAgents []AppTorrentAgent
	db.Order("id").Find(&userAgents)

	// Load credits expression
	var commonConfig CommonConfig
	db.Where("name = \"app.torrent.credits\"").First(&commonConfig)

	decoder := php_serialize.NewUnSerializer(commonConfig.Value)
	expPvalue, err := decoder.Decode()

	if err != nil {
		panic(err)
	}

	expArray, _ := expPvalue.(php_serialize.PhpArray)
	credits := make(map[int]Credit)

	for k, v := range expArray {
		vArray := v.(php_serialize.PhpArray)
		vEnabled := vArray["enabled"].(string)
		vExp := vArray["exp"].(string)

		credits[k.(int)] = Credit{enabled: vEnabled == "1", exp: vExp}
	}

	// Load logging switch
	db.Where("name = \"app.torrent.log\"").First(&commonConfig)

	// Prepare TrackerResource
	tr := &TrackerResource{setting: s, userAgents: userAgents, credits: credits, log: commonConfig.Value == "1"}

	// Initialize IRIS
	app := iris.New()
	app.OnErrorCode(iris.StatusNotFound, func(c iris.Context) {
		berror(c, "错误：Passkey 不能为空")
	})

	app.Get("/{passkey}", tr.HTTPAnnouncementHandler)
	app.Run(iris.Addr(s.Listen))
}

// HTTPAnnouncementHandler is the handler for BEP-3 Tracker GET requests.
func (tr *TrackerResource) HTTPAnnouncementHandler(c iris.Context) {
	// Get User-Agent

	userAgent := string(c.Request().UserAgent())

	// Get Passkey from url
	passkey := c.Params().Get("passkey")

	// Check parameters
	m := c.URLParams()
	requiredParams := []string{"info_hash", "peer_id", "port", "uploaded", "downloaded", "left"}
	for _, paramKey := range requiredParams {
		if _, ok := m[paramKey]; !ok {
			berror(c, fmt.Sprintf("错误：缺少参数 %s", paramKey))
			return
		}
	}

	// Get URL parameters
	event := c.URLParam("event")
	infoHash := c.URLParam("info_hash")
	peerID := c.URLParam("peer_id")
	port, _ := c.URLParamInt("port")
	uploaded, _ := c.URLParamInt("uploaded")
	downloaded, _ := c.URLParamInt("downloaded")
	left, _ := c.URLParamInt("left")

	// Get client IP
	ips := strings.Split(c.GetHeader("X-FORWARDED-FOR"), ", ")
	ip := ips[0]
	if ip == "" {
		ip = c.RemoteAddr()
	}

	if ip == "" {
		berror(c, "错误：无法获取客户端IP")
		return
	}

	// Check if User-Agent allowed
	allowed := false
	for _, v := range tr.userAgents {
		if allowed, _ = regexp.MatchString(v.AgentPattern, userAgent); allowed {
			if len(v.PeerIDPattern) > 0 {
				if allowed, _ = regexp.MatchString(v.PeerIDPattern, peerID); allowed {
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

	// Start Database connection
	db, err := gorm.Open("mysql", tr.setting.DSN)
	db.LogMode(tr.setting.Debug)

	if err != nil {
		berror(c, "错误：数据库连接失败")
		return
	}

	defer db.Close()

	// Get user info by passkey
	var user AppTorrentUser
	db.Where("passkey = ?", passkey).First(&user)

	if (AppTorrentUser{}) == user {
		berror(c, "错误：无效的 passkey，请尝试重新下载种子")
		return
	}

	// Check if BBS user existed
	var pwuser User
	db.Where("uid = ?", user.UID).First(&pwuser)

	if (User{}) == pwuser {
		berror(c, "错误：用户不存在，请尝试重新下载种子")
		return
	}

	// Check if BBS user is banned
	var userBan UserBan
	db.Where("uid = ?", user.UID).First(&userBan)

	if (UserBan{}) != userBan {
		berror(c, fmt.Sprintf("错误：用户已被封禁 %s", userBan.Reason))
		return
	}

	// Get torrent info by info_hash
	var torrent AppTorrent
	db.Where("info_hash = ?", infoHash).First(&torrent)

	if (AppTorrent{}) == torrent {
		berror(c, "错误：种子信息未注册，可能是已被删除")
		return
	}

	var bbsThread BbsThread
	db.Where("tid = ?", torrent.Tid).First(&bbsThread)

	if (BbsThread{}) == bbsThread {
		berror(c, "错误：种子不存在")
		return
	}

	if bbsThread.Disabled > 0 && bbsThread.CreatedUserid != user.UID {
		if pwuser.Groupid < 3 || pwuser.Groupid > 5 {
			berror(c, "错误：种子已删除或待审核")
			return
		}
	}

	// Log announcement
	if tr.log {
		db.Create(&AppTorrentLog{
			UID:         user.UID,
			TorrentID:   torrent.ID,
			Agent:       userAgent,
			Passkey:     passkey,
			InfoHash:    infoHash,
			PeerID:      peerID,
			IP:          ip,
			Port:        port,
			Uploaded:    uploaded,
			Downloaded:  downloaded,
			Left:        left,
			AnnouncedAt: time.Now(),
		})
	}

	// Get peers list by torrent
	torrent.Seeders = 0
	torrent.Leechers = 0
	var self AppTorrentPeer
	var peers []AppTorrentPeer
	db.Where("torrent_id = ?", torrent.ID).Find(&peers)

	i := 0
	for _, peer := range peers {
		if peer.UID == user.UID {
			// Get self from peers list by Uid
			self = peer
			peers = append(peers[:i], peers[i+1:]...)
		} else {
			// Count seeders and leechers
			if peer.Seeder {
				torrent.Seeders++
			} else {
				torrent.Leechers++
			}
		}

		i++
	}

	if (AppTorrentPeer{}) == self {
		// Create peer if not exist
		self.UID = user.UID
		self.TorrentID = torrent.ID
		self.Username = pwuser.Username
		self.IP = ip
		self.PeerID = peerID
		self.Port = port
		self.Uploaded = uploaded
		self.Downloaded = downloaded
		self.Left = left
		self.Agent = userAgent
		self.StartedAt = time.Now()
		self.LastAction = time.Now()
	}

	// Check if self is seeder
	self.Seeder = left <= 0

	if self.Seeder {
		torrent.Seeders++
	} else {
		torrent.Leechers++
	}

	// Check if peer is connectable
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		self.Connectable = false
	} else {
		self.Connectable = true
		defer conn.Close()
	}

	// Check if already started
	if self.PeerID != peerID || self.IP != ip {
		berror(c, "错误：同一种子禁止多处下载")
		return
	}

	// Get history by torrent ID and Uid
	var history AppTorrentHistory
	db.Where("torrent_id = ? AND uid = ?", torrent.ID, user.UID).First(&history)

	var rotio float64
	var uploadedAdd, downloadedAdd int

	if (AppTorrentHistory{}) == history {
		// Create history if not exist
		history = AppTorrentHistory{
			UID:        user.UID,
			TorrentID:  torrent.ID,
			Uploaded:   uploaded,
			Downloaded: downloaded,
			Left:       left,
			Leeched:    0,
			Seeded:     0,
		}

		db.Create(&history)
	} else {
		// Calculate increment
		uploadedAdd = int(math.Max(0, float64(uploaded-self.Uploaded)))
		downloadedAdd = int(math.Max(0, float64(downloaded-self.Downloaded)))

		history.Uploaded = history.Uploaded + uploadedAdd
		history.Downloaded = history.Downloaded + downloadedAdd

		history.Left = left

		if self.Seeder {
			history.Seeded += int(time.Since(self.LastAction).Seconds())
		} else {
			history.Leeched += int(time.Since(self.LastAction).Seconds())
		}

		db.Save(&history)
	}

	if len(tr.credits) > 0 {
		// Calculate rotio
		if history.Downloaded != 0 {
			rotio = math.Floor(float64(history.Uploaded/history.Downloaded*100)+0.5) / 100
		} else {
			rotio = 1
		}

		// Prepare parameters for credits calculator
		parameters := make(map[string]interface{}, 19)

		parameters["e"] = math.E
		parameters["pi"] = math.Pi
		parameters["phi"] = math.Phi

		var seeding []AppTorrentPeer
		db.Where("uid = ? AND seeder = 1", user.UID).Find(&seeding)

		var leeching []AppTorrentPeer
		db.Where("uid = ? AND seeder = 0", user.UID).Find(&leeching)

		var publishedTorrents []AppTorrent
		db.Where("Owner = ?", user.UID).Find(&publishedTorrents)

		var userData UserData
		db.Where("uid = ?", user.UID).Find(&userData)

		var windidUserData WindidUserData
		db.Where("uid = ?", user.UID).Find(&windidUserData)

		parameters["alive"] = int(time.Since(torrent.CreatedAt).Hours() / 24)
		parameters["seeders"] = torrent.Seeders
		parameters["leechers"] = torrent.Leechers
		parameters["size"] = torrent.Size
		parameters["seeding"] = len(seeding)
		parameters["leeching"] = len(leeching)
		parameters["downloaded"] = history.Uploaded
		parameters["downloaded_add"] = downloadedAdd
		parameters["uploaded"] = history.Uploaded
		parameters["uploaded_add"] = uploadedAdd
		parameters["rotio"] = rotio
		parameters["time"] = int(time.Since(self.StartedAt).Seconds())
		parameters["time_la"] = int(time.Since(self.LastAction).Seconds())
		parameters["time_leeched"] = history.Leeched
		parameters["time_seeded"] = history.Seeded
		parameters["torrents"] = len(publishedTorrents)

		// Calculate increment of credits
		for k, v := range tr.credits {
			if !v.enabled {
				continue
			}

			creditKey := fmt.Sprintf("Credit%d", k)
			parameters["credit"], _ = reflections.GetField(userData, creditKey)

			expression, _ := govaluate.NewEvaluableExpressionWithFunctions(v.exp, functions)

			delta, _ := expression.Evaluate(parameters)
			result := parameters["credit"].(float64) + delta.(float64)

			reflections.SetField(&userData, creditKey, result)
			reflections.SetField(&windidUserData, creditKey, result)
		}

		// Update credits
		db.Save(&userData)
		db.Save(&windidUserData)
	}

	// Update peer
	switch event {
	case "", "started":
		{
			self.Port = port
			self.Uploaded = uploaded
			self.Downloaded = downloaded
			self.Left = left
			self.Agent = userAgent
			self.LastAction = time.Now()

			db.Save(&self)
		}
	case "stopped":
		{
			if self.ID != 0 {
				db.Delete(&self)
			}
		}
	case "completed":
		{
			self.Port = port
			self.Uploaded = uploaded
			self.Downloaded = downloaded
			self.Left = left
			self.Agent = userAgent
			self.FinishedAt = time.Now()
			self.LastAction = time.Now()

			db.Save(&self)
		}
	default:
		{
			berror(c, "错误：客户端发送未知状态")
			return
		}
	}

	// Update torrent
	torrent.UpdatedAt = time.Now()

	db.Save(&torrent)

	// Output peers list to client
	peerList := PeerList{
		Interval:    840,
		MinInterval: 30,
		Complete:    torrent.Seeders,
		Incomplete:  torrent.Leechers,
		Peers:       peers,
	}

	buf := new(bytes.Buffer)

	bencode.Marshal(buf, peerList)

	c.Text(buf.String())
}

func berror(c iris.Context, msg string) {
	err := Error{msg}

	buf := new(bytes.Buffer)

	bencode.Marshal(buf, err)

	c.Text(buf.String())
}

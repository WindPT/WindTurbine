package main

import "time"

// Setting defines the struct for XML configuration file.
type Setting struct {
	DSN         string
	TablePrefix string
	Listen      string
	Debug       bool
}

// TrackerResource take configuration into the handler
type TrackerResource struct {
	setting    Setting
	userAgents []AppTorrentAgent
	credits    map[int]Credit
	log        bool
}

// Credit store the expressions for credits
type Credit struct {
	enabled bool
	exp     string
}

// Error for BEncode
type Error struct {
	reason string `bencode:"failure reason"`
}

// PeerList for BEncode
type PeerList struct {
	Interval    int              `bencode:"interval"`
	MinInterval int              `bencode:"min interval"`
	Complete    int              `bencode:"complete"`
	Incomplete  int              `bencode:"incomplete"`
	Peers       []AppTorrentPeer `bencode:"peers"`
}

// AppTorrent table
type AppTorrent struct {
	ID        int `gorm:"AUTO_INCREMENT;primary_key"`
	Tid       int
	InfoHash  string
	Size      int
	Leechers  int
	Seeders   int
	Owner     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AppTorrentAgent table
type AppTorrentAgent struct {
	ID            int `gorm:"AUTO_INCREMENT;primary_key"`
	Family        string
	PeerIDPattern string
	AgentPattern  string
	HTTPS         bool `gorm:"DEFAULT:0"`
	Hits          int
}

// AppTorrentHistory table
type AppTorrentHistory struct {
	ID         int `gorm:"AUTO_INCREMENT;primary_key"`
	UID        int
	TorrentID  int
	Uploaded   int
	Downloaded int
	Left       int
	Leeched    int
	Seeded     int
}

// AppTorrentLog table
type AppTorrentLog struct {
	ID          int `gorm:"AUTO_INCREMENT;primary_key"`
	UID         int
	TorrentID   int
	Agent       string
	Passkey     string
	InfoHash    string
	PeerID      string
	IP          string
	Port        int
	Uploaded    int
	Downloaded  int
	Left        int
	AnnouncedAt time.Time
}

// AppTorrentPeer table
type AppTorrentPeer struct {
	ID          int       `gorm:"AUTO_INCREMENT;primary_key" bencode:"-"`
	UID         int       `bencode:"-"`
	TorrentID   int       `bencode:"-"`
	Username    string    `bencode:"-"`
	IP          string    `bencode:"ip"`
	PeerID      string    `bencode:"peer id"`
	Port        int       `bencode:"port"`
	Uploaded    int       `bencode:"-"`
	Downloaded  int       `bencode:"-"`
	Left        int       `bencode:"-"`
	Seeder      bool      `gorm:"DEFAULT:0" bencode:"-"`
	Connectable bool      `gorm:"DEFAULT:0" bencode:"-"`
	Agent       string    `bencode:"-"`
	StartedAt   time.Time `bencode:"-"`
	FinishedAt  time.Time `bencode:"-"`
	LastAction  time.Time `bencode:"-"`
}

// AppTorrentUser table
type AppTorrentUser struct {
	UID     int `gorm:"primary_key"`
	Passkey string
}

// BbsThread table
type BbsThread struct {
	Tid           int `gorm:"AUTO_INCREMENT;primary_key"`
	Disabled      int
	CreatedUserid int
}

// CommonConfig table
type CommonConfig struct {
	Name      string
	Namespace string
	Value     string
}

// User table
type User struct {
	UID      int `gorm:"AUTO_INCREMENT;primary_key"`
	Username string
	Groupid  int
}

// UserBan table
type UserBan struct {
	ID     int `gorm:"AUTO_INCREMENT;primary_key"`
	UID    int
	Reason string
}

// UserData table
type UserData struct {
	UID     int `gorm:"AUTO_INCREMENT;primary_key"`
	Credit1 float64
	Credit2 float64
	Credit3 float64
	Credit4 float64
	Credit5 float64
	Credit6 float64
	Credit7 float64
	Credit8 float64
}

// WindidUserData table
type WindidUserData UserData

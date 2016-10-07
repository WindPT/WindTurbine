package main

import (
    "time"
    "github.com/jinzhu/gorm"
)

type Setting struct {
    DSN         string
    TablePrefix string
    Listen      string
    Debug       bool
}

type TrackerResource struct {
    db          *gorm.DB
    user_agents []AppTorrentAgent
}

/*
 * Structs for BEncode
 */

type Error struct {
    reason string `bencode:"failure reason"`
}

type PeerList struct {
    Interval    int              `bencode:"interval"`
    MinInterval int              `bencode:"min interval"`
    Complete    int              `bencode:"complete"`
    Incomplete  int              `bencode:"incomplete"`
    Peers       []AppTorrentPeer `bencode:"peers"`
}

/*
 * Structs for DB
 */

type AppTorrent struct {
    Id           int              `gorm:"AUTO_INCREMENT;primary_key"`
    Tid          int
    InfoHash     string
    Size         int
    Leechers     int
    Seeders      int
    Owner        int
    UpdatedAt    time.Time
    CreatedAt    time.Time
}

type AppTorrentAgent struct {
    Id              int   `gorm:"AUTO_INCREMENT;primary_key"`
    Family          string
    PeerIdPattern   string
    AgentPattern    string
    Https           bool   `gorm:"DEFAULT:0"`
    Hits            int
}

type AppTorrentHistory struct {
    Id             int `gorm:"AUTO_INCREMENT;primary_key"`
    Uid            int
    TorrentId      int
    Uploaded       int
    UploadedLast   int
    Downloaded     int
    DownloadedLast int
    Left           int
    State          string
}

type AppTorrentPeer struct {
    Id          int       `gorm:"AUTO_INCREMENT;primary_key" bencode:"-"`
    TorrentId   int       `bencode:"-"`
    Uid         int       `bencode:"-"`
    Ip          string    `bencode:"ip"`
    PeerId      string    `bencode:"peer id"`
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

type AppTorrentUser struct {
    Uid          int     `gorm:"primary_key"`
    Passkey      string
    UploadedMo   int
    DownloadedMo int
}

type User struct {
    Uid      int `gorm:"AUTO_INCREMENT;primary_key"`
    Username string
    Groupid  int
}

type UserBan struct {
    Id     int `gorm:"AUTO_INCREMENT;primary_key"`
    Uid    int
    Reason string
}

type BbsThread struct {
    Tid           int `gorm:"AUTO_INCREMENT;primary_key"`
    Disabled      int
    CreatedUserid int
}

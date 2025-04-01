package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dushixiang/next-terminal-export/model"
	"github.com/dushixiang/next-terminal-export/utils"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Sqlite struct {
	File string `yaml:"file"`
}

type Mysql struct {
	Hostname string `yaml:"hostname"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type Config struct {
	DB     string `yaml:"db"`
	Sqlite Sqlite `yaml:"sqlite"`
	Mysql  Mysql  `yaml:"mysql"`
}

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "c", "config.yml", "config path")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func readConfig() (*Config, error) {
	file, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file failed: %v", err)
	}
	conf := new(Config)
	err = yaml.Unmarshal(file, conf)
	if err != nil {
		return nil, fmt.Errorf("unmarshal config file failed: %v", err)
	}
	return conf, nil
}

func initDB(conf *Config) (*gorm.DB, error) {
	switch conf.DB {
	case "sqlite":
		return initSqlite(conf.Sqlite)
	case "mysql":
		return initMysql(conf.Mysql)
	default:
		return nil, fmt.Errorf("not support db: %v", conf.DB)
	}
}

func initMysql(conf Mysql) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", conf.Username, conf.Password, conf.Hostname, conf.Port, conf.Database)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql failed: %v", err)
	}
	return db, nil
}

func initSqlite(conf Sqlite) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(conf.File), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sqlite failed: %v", err)
	}
	return db, nil
}

func main() {
	log.Println("reading config...")
	conf, err := readConfig()
	if err != nil {
		log.Fatalf("read config failed: %v", err)
	}
	log.Printf("config: %+v", conf)
	log.Println("init db...")
	db, err := initDB(conf)
	if err != nil {
		log.Fatalf("init db failed: %v", err)
	}
	log.Println("init db success")
	err = export(db)
	if err != nil {
		log.Fatalf("export db failed: %v", err)
	}
}

type Backup struct {
	Users            []model.User             `json:"users"`
	UserGroups       []model.UserGroup        `json:"user_groups"`
	Storages         []model.Storage          `json:"storages"`
	Strategies       []model.Strategy         `json:"strategies"`
	AccessSecurities []model.AccessSecurity   `json:"access_securities"`
	AccessGateways   []model.AccessGateway    `json:"access_gateways"`
	Commands         []model.Command          `json:"commands"`
	Credentials      []model.Credential       `json:"credentials"`
	Assets           []map[string]interface{} `json:"assets"`
	Jobs             []model.Job              `json:"jobs"`
}

func export(db *gorm.DB) (err error) {
	log.Println("exporting...")
	var (
		users          []model.User
		userGroups     []model.UserGroup
		storages       []model.Storage
		strategies     []model.Strategy
		jobs           []model.Job
		accessGateways []model.AccessGateway
		commands       []model.Command
		credentials    []model.Credential
		assets         []model.Asset
	)

	if db.Find(&users).Error != nil {
		return err
	}
	for i := range users {
		users[i].Password = ""
	}
	if db.Find(&userGroups).Error != nil {
		return err
	}
	if len(userGroups) > 0 {
		for i := range userGroups {
			var members []string
			err = db.Table("user_group_members").Select("user_id").Where("user_group_id = ?", userGroups[i].ID).Find(&members).Error
			if err != nil {
				return err
			}
			userGroups[i].Members = members
		}
	}

	if db.Find(&storages).Error != nil {
		return err
	}

	if db.Find(&strategies).Error != nil {
		return err
	}
	if db.Find(&jobs).Error != nil {
		return err
	}
	if db.Find(&accessGateways).Error != nil {
		return err
	}
	if db.Find(&commands).Error != nil {
		return err
	}
	if db.Find(&credentials).Error != nil {
		return err
	}
	if len(credentials) > 0 {
		for i := range credentials {
			credentials[i].Password = utils.MustDecrypt(credentials[i].Password)
			credentials[i].PrivateKey = utils.MustDecrypt(credentials[i].PrivateKey)
			credentials[i].Passphrase = utils.MustDecrypt(credentials[i].Passphrase)
		}
	}
	if db.Table("assets").Find(&assets).Error != nil {
		return err
	}
	var assetMaps = make([]map[string]interface{}, 0)
	if len(assets) > 0 {
		for i := range assets {
			assets[i].Password = utils.MustDecrypt(assets[i].Password)
			assets[i].PrivateKey = utils.MustDecrypt(assets[i].PrivateKey)
			assets[i].Passphrase = utils.MustDecrypt(assets[i].Passphrase)

			asset := assets[i]
			attributeMap, err := findAssetAttrMapByAssetId(db, asset.ID)
			if err != nil {
				return err
			}
			itemMap := utils.StructToMap(asset)
			for key := range attributeMap {
				itemMap[key] = attributeMap[key]
			}
			itemMap["created"] = asset.Created.Format("2006-01-02 15:04:05")
			assetMaps = append(assetMaps, itemMap)
		}
	}

	backup := Backup{
		Users:          users,
		UserGroups:     userGroups,
		Storages:       storages,
		Strategies:     strategies,
		Jobs:           jobs,
		AccessGateways: accessGateways,
		Commands:       commands,
		Credentials:    credentials,
		Assets:         assetMaps,
	}
	// roles
	indent, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile("backup.json", indent, 0666)
	if err != nil {
		return err
	}
	log.Println("export success!!! 🎉🎉🎉")
	return nil
}

func findAssetAttrMapByAssetId(db *gorm.DB, assetId string) (map[string]any, error) {
	var attrs []model.AssetAttribute
	err := db.Table("asset_attributes").Where("asset_id = ?", assetId).Find(&attrs).Error
	if err != nil {
		return nil, err
	}
	var attrMap = make(map[string]any)
	for _, attr := range attrs {
		attrMap[attr.Name] = attr.Value
	}
	return attrMap, nil
}

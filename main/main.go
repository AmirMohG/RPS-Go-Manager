package main

import (
	_ "database/sql"
	"errors"
	"fmt"
	_ "net/http"
	"os"
	"os/exec"
	_ "strconv"
	"strings"
	_ "time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"rps.local/date"
)

type DB struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
}
type reqBody struct {
	Username     string  `json:"username"`
	Password     string  `json:"password"`
	Vmid         int     `json:"vmid"`
	Duration     int     `json:"duration"`
	Ram          float32 `json:"ram"`
	Cpu          float32 `json:"cpu"`
	Price        int     `json:"price"`
	Command      string  `json:"command"`
	AllowedPorts int     `json:"allowed_port" gorm:"column:allowed_port;not null"`
	CurrentPorts int     `json:"current_port" gorm:"column:current_port;not null"`
}
type User struct {
	Username     string `json:"username" gorm:"primaryKey" gorm:"column:username;not_null;type:varchar;size:32;unique_index"`
	Password     string `json:"password" gorm:"column:password;not null;size:64"`
	Date_created string `json:"date_created" gorm:"column:date_created;not null;size:20"`
}

type VM struct {
	Vmid         int     `json:"vmid" gorm:"primaryKey" gorm:"column:vmid;not null`
	Owner        string  `json:"owner" gorm:"column:owner;not null;references:users(username)"`
	Expiry       string  `json:"expiry" gorm:"column:expiry;not null"`
	Ram          float32 `json:"ram" gorm:"column:ram;not null"`
	Cpu          float32 `json:"cpu" gorm:"column:cpu;not null"`
	Price        int     `json:"price" gorm:"column:price;not null"`
	AllowedPorts int     `json:"allowed_port" gorm:"column:allowed_port;not null"`
	CurrentPorts int     `json:"current_port" gorm:"column:current_port;not null"`
}

type Subdomains struct {
	Domain string `json:"domain" gorm:"primaryKey" gorm:"column:domain;not null`
	Vmid   int    `json:"vmid" gorm:"column:vmid;not null;size:10;references:users(username)"`
	Port   int    `json:"port" gorm:"column:port;not null;size:10"`
	Expiry string `json:"expiry" gorm:"column:expiry;not null;size:20"`
}

var db *gorm.DB

func main() {
	godotenv.Load(".env")
	APP_HOST := os.Getenv("APP_HOST")
	APP_PORT := os.Getenv("APP_PORT")
	ADDR := fmt.Sprintf(APP_HOST + ":" + APP_PORT)
	db = InitGormMysql(db)

	db.AutoMigrate(&VM{})
	db.AutoMigrate(&User{})
	db.AutoMigrate(&Subdomains{})
	router := gin.Default()
	router.GET("/test", func(res *gin.Context) {
		res.JSON(200, gin.H{
			"user": "pong",
		})
	})
	router.POST("/api/userExists", func(c *gin.Context) {
		var user User
		err := c.Bind(&user)
		if err != nil {
			c.JSON(400, gin.H{"Error": "Bad request"})
		} else if userExists(&user) {
			c.JSON(200, gin.H{"status": "True"})
		} else {
			c.JSON(404, gin.H{"status": "False"})
		}
	})
	router.POST("/api/register", func(c *gin.Context) {
		var user User
		err := c.Bind(&user)
		if err != nil {
			c.JSON(400, gin.H{"Error": "Bad request"})
		} else if !userExists(&user) {
			user.Password = hash(user.Password)
			code, err := register(&user)
			c.JSON(code, gin.H{"Error": err.Error()})
		} else {
			c.JSON(403, gin.H{"Error": "Bad Credentials"})
		}

	})
	router.POST("/api/buy", func(c *gin.Context) {
		var user User
		var vm VM
		var req reqBody
		err := c.Bind(&req)
		if err != nil {
			c.JSON(400, gin.H{"Error": "Bad request"})
		} else {
			user.Username = req.Username
			user.Password = req.Password
			vm.Vmid = req.Vmid
			vm.Owner = req.Username
			vm.Ram = req.Ram
			vm.Cpu = req.Cpu
			vm.Price = req.Price
			duration := req.Duration

			if userExists(&user) && paid() {
				code, err := buyVM(&user, &vm, duration)
				c.JSON(code, gin.H{"Error": err.Error()})
			} else {
				c.JSON(403, gin.H{"Error": "Bad credentials"})
			}
		}

	})
	router.POST("/api/command", func(c *gin.Context) {
		var user User
		var vm VM
		var req reqBody
		err := c.Bind(&req)
		if err != nil {
			c.JSON(400, gin.H{"Error": "Bad request"})
		} else {
			user.Username = req.Username
			vm.Vmid = req.Vmid
			if !authorized(user, vm) {
				c.JSON(401, gin.H{"Error": "Unauthorized"})
			} else {
				res, err := doCommand(req.Command)
				if err != nil {
					c.JSON(500, gin.H{"Error": err})
				} else {
					c.JSON(200, gin.H{"Result": res})
				}
			}
		}

	})
	router.Run(ADDR)

}
func InitGormMysql(conn *gorm.DB) *gorm.DB {
	var err error
	dbConf := DB{
		Username: os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASS"),
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     os.Getenv("MYSQL_PORT"),
		Database: os.Getenv("MYSQL_DB"),
	}
	fmt.Println("CONFIG", dbConf)
	conn, err = gorm.Open(mysql.New(mysql.Config{
		DSN: fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbConf.Username,
			dbConf.Password,
			dbConf.Host,
			dbConf.Port,
			dbConf.Database,
		),
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	}))
	if err != nil {
		panic(err)
	}
	return conn
}
func register(user *User) (int, error) {
	if len(user.Username) < 32 && len(user.Username) > 3 && len(user.Password) > 8 {
		user.Username = strings.ToLower(user.Username)
		var date_created date.Date
		date.Now(&date_created)
		user.Date_created = date.ToString(date_created)
		res := db.Create(&user)
		if res.Error != nil {
			panic(res.Error)
		}
	} else {
		return 403, errors.New("Bad credentials")
	}
	return 200, nil

}

func userExists(user *User) bool {

	res := db.First(&user, "username = ?", strings.ToLower(user.Username))
	fmt.Println()
	if res.Error != nil {
		return false
	} else {
		return true
	}
}
func hash(password string) string {
	bytePass := []byte(password)

	hashedPassword, err := bcrypt.GenerateFromPassword(bytePass, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashedPassword[:])
}
func buyVM(user *User, vm *VM, duration int) (int, error) {
	if vm.AllowedPorts == 0 {
		vm.AllowedPorts = 3
	}
	if vm.Vmid == 0 || vm.Ram < 100 || vm.Cpu < 1 || vm.Price < 5 {
		return 403, errors.New("Bad input")
	}
	var expiry date.Date
	date.Now(&expiry)
	date.IncMonth(&expiry, duration)
	vm.Expiry = date.ToString(expiry)
	res := db.Create(vm)
	if res.Error != nil {
		return 500, errors.New("Failed to create")
	} else {
		return 200, errors.New("Success")
	}
}

func authorized(user User, vm VM) bool {
	if !userExists(&user) {
		return false
	}
	res := db.First(&vm, "owner = ?", strings.ToLower(user.Username))
	if res.Error != nil {
		return false
	} else {
		return true
	}

	return false
}
func paid() bool {
	return true
}

//	func restartVM(user User, vm VM) string{
//	    if authorized(user, vm) {
//	        res, err := doCommand("qm restart " + vm.Vmid)
//	        if err != nil {
//	            panic(err)
//	            return err.Error()
//	        }
//	        return string(res[:])
//	    } else {
//	        return "401"
//	    }
//	}
func doCommand(cmd string) (string, error) {
	args := strings.Fields(cmd)
	stdout, stderr := exec.Command(args[0], args[1:]...).Output()
	if stderr != nil {
		return "", stderr
	}
	return string(stdout), nil
}

// func addSubdomain(user User, vmid)

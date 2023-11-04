package database

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rodrikv/openai_proxy/utils"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var Db *gorm.DB

type App struct {
	DB *gorm.DB
}

func init() {
	InitDatabase()
}

func InitEndpoints() {
	data, err := os.ReadFile("config/endpoints.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Define the struct to hold the JSON data
	type EndpointData struct {
		Endpoints []Endpoint `json:"endpoints"`
	}

	// Unmarshal JSON data into EndpointData struct
	var endpointData EndpointData
	err = json.Unmarshal(data, &endpointData)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	// Insert endpoints into the database
	for _, e := range endpointData.Endpoints {
		m := Model{}
		nm := make([]Model, 0)
		for _, mname := range e.Models {
			err := m.Get(mname.Name)
			if err == nil {
				nm = append(nm, m)
			}
		}
		err = e.Create(e.Name, e.Url, e.Token, e.Concurrent, e.IsActive, &e.RPM, &e.RPD, &nm)
	}

	if err != nil {
		logrus.Error("Error inserting endpoints:", err)
	}
}

func InitUsers() {
	data, err := os.ReadFile("config/users.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	type U struct {
		User      User     `json:"user"`
		Endpoints []string `json:"endpoints"`
	}

	// Define the struct to hold the JSON data
	type UserData struct {
		Users []U `json:"users"`
	}

	// Unmarshal JSON data into UserData struct
	var userData UserData
	err = json.Unmarshal(data, &userData)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	for _, u := range userData.Users {
		var ees []Endpoint
		es := u.Endpoints

		err = InsertUser(u.User)

		if err != nil {
			continue
		}

		Db.Where("name = ?", u.User.Name).First(&u.User)

		if len(es) == 0 {
			Db.Find(&ees)
		} else {
			for _, e := range es {
				var ee Endpoint
				Db.Where("name = ?", e).First(&ee)
				ees = append(ees, ee)
			}
		}

		for _, ee := range ees {
			ee.AddUser(&u.User)
		}
	}
}

type ModelData struct {
	Models []Model `json:"models"`
}

func createModels(models []Model) []Model {
	createdModels := make([]Model, 0)

	for _, m := range models {
		newModel := Model{}
		var ms []Model
		if m.SubModels != nil {
			ms = createModels(m.SubModels)
		}

		newModel.Create(m.Name, ms)

		fmt.Println("Creating model:", ms)
		fmt.Println("Created model:", newModel)

		createdModels = append(createdModels, newModel)
	}

	return createdModels
}

func InitModels() {
	data, err := os.ReadFile("config/models.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	// Unmarshal JSON data into ModelData struct
	var modelData ModelData
	err = json.Unmarshal(data, &modelData)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	// Insert models into the database
	createModels(modelData.Models)

	if err != nil {
		logrus.Error("Error inserting models:", err)
	}
}

func GetDatabase(name, url string) gorm.Dialector {
	if name == "sqlite" {
		return sqlite.Open(url)
	} else if name == "mysql" {
		return mysql.Open(url)
	}
	return nil
}

func InitDatabase() {
	dbname := utils.Getenv("DB_NAME", "sqlite")
	dburl := utils.Getenv("DB_URL", "admin.db")

	db, dberr := gorm.Open(GetDatabase(dbname, dburl), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "proxy_",
		},
	})

	if dberr != nil {
		panic("failed to connect to database")
	}
	Db = db
	_, sqldberr := db.DB()
	if sqldberr != nil {
		panic("failed to get db")
	}

	migrationerr := db.AutoMigrate(&User{}, &Endpoint{}, &UserEndpoint{}, &Model{}, &EndpointModelUsage{})
	// defer sqlDB.Close()

	if migrationerr != nil {
		panic("failed to migrate")
	}

	InitModels()
	InitEndpoints()
	InitUsers()
}

func ResetUsageToday() {
	var users []User
	result := Db.Find(&users)
	if result.Error != nil {
		fmt.Println(result.Error)
	}
	for _, user := range users {
		user.UsageToday = 0
		UpdateUserUsageToday(user)
	}
}

func ResetEndpointUsage() {
	var endpoints []Endpoint
	result := Db.Find(&endpoints)
	if result.Error != nil {
		fmt.Println(result.Error)
	}
	for _, endpoint := range endpoints {
		endpoint.ResetEndpointUsage()
	}
}

func ResetEndpointDailyUsage() {
	var endpoints []Endpoint
	result := Db.Find(&endpoints)
	if result.Error != nil {
		fmt.Println(result.Error)
	}
	for _, endpoint := range endpoints {
		endpoint.ResetEndpointDailyUsage()
	}
}

func ResetRequestCount() {
	var users []User
	result := Db.Find(&users)
	if result.Error != nil {
		fmt.Println(result.Error)
	}
	for _, user := range users {
		user.ResetRequestCount()
	}
}

package controllers

import (
	"encoding/xml"
	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
	"io/ioutil"
	"log"
	"net/http"
	"sell/models"
	"sell/util"
	"strings"
	"time"
)

type SdnList struct {
	XMLName  xml.Name `xml:"sdnList"`
	SdnEntry []struct {
		UID       uint   `xml:"uid"`
		FirstName string `xml:"firstName"`
		LastName  string `xml:"lastName"`
		SdnType   string `xml:"sdnType"`
	} `xml:"sdnEntry"`
}

var name models.Name
var names []models.Name

var busy = false

func Update(c *gin.Context) {
	// locking process
	if busy {
		c.JSON(http.StatusOK, gin.H{
			"error": "Process is already in progress",
		})
		return
	}

	defer func() {
		if err := recover(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"result": false,
				"info":   "service unavailable",
				"code":   503})
			busy = false
		}
	}()

	busy = true

	url := "https://www.treasury.gov/ofac/downloads/sdn.xml"

	req, _ := http.NewRequest("GET", url, nil)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatal(err)
		return
	}

	list := new(SdnList)
	err = xml.Unmarshal([]byte(body), list)
	if err != nil {
		log.Fatal(err)
		return
	}
	db := util.DB

	result := db.Find(&names)
	rows, _ := result.Rows()
	defer rows.Close()
	var nameCompare []models.Name
	var ids []uint

	for rows.Next() {
		db.ScanRows(rows, &name)
		nameCompare = append(nameCompare, models.Name{FirstName: name.FirstName, LastName: name.LastName, UID: name.UID})
		ids = append(ids, name.UID)
	}

	for _, name := range list.SdnEntry {
		if name.SdnType != "Individual" {
			continue
		}

		nameSave := models.Name{FirstName: name.FirstName, LastName: name.LastName, UID: name.UID, FullName: name.FirstName + " " + name.LastName}

		// if we have the same record already in the DB, we skip
		if slices.Contains(nameCompare, nameSave) {
			continue
		}
		// if we have a record with the same UID - we update
		if slices.Contains(ids, nameSave.UID) {
			// will be escaped
			// https://gorm.io/docs/security.html
			util.DB.Exec("UPDATE names set first_name = ? , last_name = ? , full_name = ? , updated_at = ? where uid = ? ", name.FirstName, name.LastName, name.FirstName+" "+name.LastName, time.Now(), name.UID)
		} else {
			// otherwise we insert
			util.DB.Create(&nameSave)
		}
	}

	busy = false

	c.JSON(http.StatusOK, gin.H{
		"code":   200,
		"info":   "",
		"result": true,
	})

}

func State(c *gin.Context) {

	if busy {
		c.JSON(http.StatusOK, gin.H{
			"info":   "updating",
			"result": false,
		})
		return
	}

	result := util.DB.First(&name)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusOK, gin.H{
			"info":   "empty",
			"result": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"info":   "ok",
		"result": true,
	})
	return
}

type NameResource struct {
	UID       uint
	FirstName string
	LastName  string
}

func GetNames(c *gin.Context) {
	nameParam := c.Query("name")
	typeParam := c.Query("type")
	db := util.DB
	var namesSlice = []NameResource{}
	if strings.ToLower(typeParam) != "weak" {
		result := db.Where("full_name = ?", nameParam).Find(&names)
		rows, _ := result.Rows()
		defer rows.Close()
		for rows.Next() {
			db.ScanRows(rows, &name)
			namesSlice = append(namesSlice, NameResource{LastName: name.LastName, FirstName: name.FirstName, UID: name.UID})
		}

		c.JSON(http.StatusOK, namesSlice)
		return
	} else {

		explodedNameParts := strings.Split(nameParam, " ")
		semiCompleted := db.Where("0")
		for _, s := range explodedNameParts {
			semiCompleted = semiCompleted.Or("first_name LIKE ? ", "%"+s+"%").Or("last_name LIKE ? ", "%"+s+"%")
		}
		result := semiCompleted.Find(&names)
		rows, _ := result.Rows()
		defer rows.Close()
		for rows.Next() {
			db.ScanRows(rows, &name)
			namesSlice = append(namesSlice, NameResource{LastName: name.LastName, FirstName: name.FirstName, UID: name.UID})
		}
		c.JSON(http.StatusOK, namesSlice)
		return

	}

}

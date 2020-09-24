package main

import (
	"ProjectMongoClient"
	classes "activities"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"

	//"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

var session *ProjectMongoClient.DBSession

func main() {



	var err error
	tablesMap := make(map[string]string)
	tablesMap["user"] ="userstable"
	tablesMap["activities"] = "activitiestable"
	tablesMap["orders"] = "orderstable"
	tablesMap["announcements"] = "announcementstable"

	session, err = ProjectMongoClient.NewSession("toursdb",tablesMap)
	if err != nil{
		fmt.Println(err.Error())
		return
	}
	defer session.Close()

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("./style","./templates")
	err = RouterGroupsInit(router)
	if err != nil{
		panic(err)
	}
}


func RouterGroupsInit(router *gin.Engine) error{

	data := router.Group("/data")
	{
		data.POST("/show", Show)
	}
	auth := router.Group("/auth")
	{
		//auth.Use(CheckUserStatus)
		auth.GET("/regform",RegForm)
		auth.GET("/signform",SignForm)

	}



	user := router.Group("/user")
	{
		user.Use(CheckUserTokenValidation)
		user.GET("/signin",SignIn)
		user.POST("/signin",SignIn)
		user.GET("/signup",SignUp)
		user.POST("/signup",SignUp)
		user.GET("/account",Account)
		user.GET("/trash",Trash)
		user.GET("/announcements",Announcements)
	}
	author := router.Group("/author")
	{
		author.GET("/page",AuthorPage)
		author.POST("/addannouncement",AddAnnouncement)
	}
	admin := router.Group("/admin")
	{
		admin.GET("/page",AdminPage)
	}
	err := router.Run(":8080")
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func RegForm(c *gin.Context)  {
	c.HTML(http.StatusOK, "registration.html", nil)
}
func SignForm(c *gin.Context)  {
	c.HTML(http.StatusOK, "signup.html", nil)
}

func AuthorPage(c *gin.Context)  {
	c.HTML(http.StatusOK, "author.html", nil)
}
func Account(c *gin.Context)  {
	var selector map[string]interface{}
	announcements,err := session.Read(selector,"announcements")
    if err != nil{
    	fmt.Println(err.Error())
    	c.String(400,err.Error())
		return
	}
	c.HTML(http.StatusOK, "account.html", announcements)
}
func AdminPage(c *gin.Context)  {
	c.HTML(http.StatusOK, "adminpage.html", nil)
}

func Show(c *gin.Context)  {
	var data interface{}
	c.BindJSON(&data)
	c.JSON(200, data)
	fmt.Println("DATA",data)
}

//middleware token authorisation for user
func CheckUserTokenValidation(c *gin.Context) {
	_, err1 := c.Cookie("token")
	status,err2 := c.Cookie("status")
	if err1 != nil || err2 != nil || status != "user" {
		c.HTML(200, "signup.html", gin.H{
			"title": "authorisation",
			"error1":err1.Error(),
			"error2":err2.Error(),
		})
		return
	}
	return
}

//middleware token authorisation for user
func CheckUserStatus(c *gin.Context) {
	status, err := c.Cookie("status")
	if err!=nil {
		c.HTML(200, "signin.html", gin.H{
			"title": "authorisation",
			"status":status,
		})
		return
	}
	return
}

func SignIn(context *gin.Context) {
	contentType := context.GetHeader("content-type")
	if contentType == "application/json" {
		mymap := make(map[string]string)
		err := context.BindJSON(&mymap)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if login, ok := mymap["login"]; ok {
			if password, ok := mymap["password"]; ok {
				if email, ok := mymap["email"]; ok {
					user, err := session.CheckUserInDB(login, email, password,"user")
					if err != nil {
						fmt.Println(err.Error())
						context.String(http.StatusNoContent, "user already exists", err.Error())
						return
					}
					err = session.Insert(user,"users")
					if err != nil {
						fmt.Println(err.Error())
						context.String(http.StatusInternalServerError, "err user add ", err.Error())
						return
					}
					fmt.Println("Sucsesfully added User ", login, "email: ", user.Email)
					context.String(http.StatusOK, "Sucsesfully added User", login)
					return
				} else {
					context.String(http.StatusAccepted, "no email info")
					return
				}
			} else {
				context.String(http.StatusAccepted, "no password info")
				return
			}
		} else {
			context.String(http.StatusAccepted, "no login info")
			return
		}
		context.String(http.StatusNoContent, "sth wrong")
		return
	} else {
		login := context.PostForm("login")
		password := context.PostForm("password")
		email := context.PostForm("email")
		status := "user"
		if login != "" && password != "" && email != "" {
			user, err := session.CheckUserInDB(login, email, password,status)
			if err != nil {
				fmt.Println(err.Error())
				context.String(http.StatusNoContent, "user already exists ", err.Error())
				return
			}
			//fmt.Println("user:", user)
			err = session.Insert(user,"user")
			if err != nil {
				fmt.Println(err.Error())
				context.String(http.StatusInternalServerError, "err user add ", err.Error())
				return
			}
			fmt.Println("Sucsesfully added User ", login, "id ", user.ID)
			context.HTML(200,"signup.html",user)
		} else {
			fmt.Println("no password or login or email info")
		}
	}
	return
}

func SignUp(context *gin.Context) {

		login,ok1:= context.GetPostForm("login")
		password,ok2:= context.GetPostForm("password")
		fmt.Println(login,password)
		if ok1 && ok2 {
			user, err := session.CheckUserPassword(login, password)
			if err != nil {
				fmt.Println(err.Error())
				context.String(http.StatusOK, err.Error())
				return
			}
			//sending token as a cookie
			tokenstring := user.GetToken()
			context.SetCookie("token", tokenstring, 5*60, "/", "localhost", false, false)
			context.SetCookie("login", user.Username, 5*60, "/", "localhost", false, false)
			context.SetCookie("status", "user", 5*60, "/", "localhost", false, false)
			//return HTML
			context.HTML(http.StatusOK, "account.html", bson.M{
				"Username":user.Username,
				"Email":user.Email,
			})
			return
		} else {
			fmt.Println("no password or email")
			context.String(http.StatusBadRequest, "no password or email")
			return
		}
}

func Trash(c *gin.Context)  {
	var selector = make(map[string]interface{})
	var orders []interface{}
	var err error
	login, err := c.Cookie("email")
	if err!=nil{
		fmt.Println(err.Error())
		c.String(400,err.Error())
		return
	}
	selector["user_login"] = login
	orders, err = session.Read(selector,"announcements")
	order := orders[0].(classes.Order)

	if err != nil {
		fmt.Println(err.Error())
		c.String(400, err.Error())
		return
	}
	c.JSON(200,order.OrderList)
}

func Announcements(c *gin.Context)  {
	var selector = make(map[string]interface{})
	var announcements  = make([]classes.Announcement,1)
	data, err := session.Read(selector,"announcements")
	fmt.Println(data)

	fmt.Println("hd",announcements)
	if err != nil {
		fmt.Errorf(err.Error())
		c.String(400, err.Error())
		return
	}
	c.String(200,announcements[0].Title)
	c.JSON(200,announcements)
}

func AddAnnouncement(c *gin.Context)  {
     var data map[string]string
     err := c.BindJSON(&data)
     if err != nil{
     	fmt.Println(err.Error())
     	return
	 }

	 login ,err := c.Cookie("login")
	 if err!= nil{
	  	fmt.Println(err.Error())
		  return
	 }
	 price, err := strconv.ParseFloat(data["price"],64)
	 if err!=nil{
	 	fmt.Println(err.Error())
		return
	 }
	 id := primitive.NewObjectID()
     announcement := classes.Announcement{
     	ID: id,
     	IDString: id.String(),
     	Title: data["title"],
     	AuthorLogin: login,
     	StartWeekDays: strings.Split(data["start_days"]," "),
     	PhoneNumber: data["phone_number"],
     	Activity:classes.Activity{
     		Name: data["name"],
     		Type: data["type"],
     		Price: price,
     		Description: data["description"],
		},
	 }
	 err = session.Insert(announcement,"announcements")
     if err != nil{
     	fmt.Println(err.Error())
     	c.String(400,err.Error())
     	return
	 }
     fmt.Println(announcement)
}



package main

import (
	"ProjectMongoClient"
	classes "activities"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strconv"
	"strings"
	//"unicode"

	//"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

var session *ProjectMongoClient.DBSession

func StartDB() error  {
	var err error
	tablesMap := make(map[string]string)
	tablesMap["user"] ="userstable"
	tablesMap["activities"] = "activitiestable"
	tablesMap["orders"] = "orderstable"
	tablesMap["announcements"] = "announcementstable"
    tablesMap["coments"] = "comentstable"
	session, err = ProjectMongoClient.NewSession("toursdb",tablesMap)
	if err != nil{
		return err
	}
	return nil
}

func main() {
	err := StartDB()
	if err != nil{
		fmt.Println(err.Error())
		return
	}
	defer session.Close()
	router := gin.Default()

	err = RouterGroupsInit(router)
	if err != nil{
		panic(err)
	}
}


func RouterGroupsInit(router *gin.Engine) error{

	//router.LoadHTMLGlob("./js_codes/*")

	//router.LoadHTMLGlob("./css_styles/*")
	router.LoadHTMLGlob("./templates/*")
	router.Static("./style","C:/Users/dmitr/go/src/newServer/css_style")
	//router.Static("./scripts","C:/Users/dmitr/go/src/newServer/js_codes")
	router.Use(cors.Default())
	data := router.Group("/data")
	{
		data.POST("/show", Show)
		data.GET("/announcements", Announcements)
		data.POST("/announcements", Announcements)
	}
	auth := router.Group("/auth")
	{
		//auth.Use(CheckUserStatus)
		auth.GET("/announcementinfo",AnnouncementInfo)
		auth.POST("/announcementinfo",AnnouncementInfo)
		auth.GET("/regform",RegForm)
		auth.GET("/signform",SignForm)
		auth.GET("/signup",SignUp)
		auth.POST("/signup",SignUp)
		auth.GET("/signin",SignIn)
		auth.POST("/signin",SignIn)
	}



	user := router.Group("/user")
	{
		user.Use(CheckUserTokenValidation)

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
	c.HTML(http.StatusOK, "signup.html", nil)
}
func SignForm(c *gin.Context)  {
	c.HTML(http.StatusOK, "signin.html", nil)
}

func AuthorPage(c *gin.Context)  {
	c.HTML(http.StatusOK, "author.html", nil)
}
func Account(c *gin.Context)  {
	var selector map[string]interface{}
	announcements,err := session.ReadAnnouncements(selector)
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
//func CheckUserStatus(c *gin.Context) {
//	status, err := c.Cookie("status")
//	if err!=nil {
//		c.HTML(200, "signin.html", gin.H{
//			"title": "authorisation",
//		})
//		return
//	}
//	if status == "user"{
//
//		return
//	}
//	if status == "author"{
//		return
//	}
//	if status =="admin"{
//		return
//	}
//	return
//}

func SignUp(context *gin.Context) {

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
			context.HTML(200,"signin.html",user)
		} else {
			fmt.Println("no password or login or email info")
		}
	return
}

func SignIn(context *gin.Context) {

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
			context.SetCookie("status", user.Status, 5*60, "/", "localhost", false, false)
			//return HTML
			if user.Status == "user"{
			context.HTML(http.StatusOK, "userAccount.html", bson.M{
				"Username":user.Username,
				"Email":user.Email,
			})
				return
			}
			if user.Status == "author"{
				context.HTML(http.StatusOK, "authorAccount.html", bson.M{
					"Username":user.Username,
					"Email":user.Email,
				})
				return
			}
			if user.Status == "admin"{
				context.HTML(http.StatusOK, "adminPage.html", bson.M{
					"Username":user.Username,
					"Email":user.Email,
				})
				return
			}else {
				context.String(400,"No status info for user "+user.Username)
			}
			return
		} else {
			fmt.Println("no password or email")
			context.String(http.StatusBadRequest, "no password or email")
			return
		}
}

func Trash(c *gin.Context)  {
	var selector = make(map[string]interface{})
	var err error
	login, err := c.Cookie("email")
	if err!=nil{
		fmt.Println(err.Error())
		c.String(400,err.Error())
		return
	}
	selector["user_login"] = login
	order, err := session.ReadOrder(selector)
	if err != nil {
		fmt.Println(err.Error())
		c.String(400, err.Error())
		return
	}
	c.JSON(200,order.ActivityList)
}

func Announcements(c *gin.Context)  {
	var selector = make(map[string]interface{})
	//var announcements  []classes.Announcement
	//var err error
	announcements, err := session.ReadAnnouncements(selector)
	if err != nil {
		fmt.Println(err.Error())
		c.String(400, err.Error())
		return
	}
	out,_ :=json.MarshalIndent(announcements,"  ","   ")
    fmt.Println(announcements[0].Title)
	fmt.Println(string(out))
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
	 title ,ok :=data["title"]
	 if !ok{
	 	fmt.Println("err no title")
		 return
	 }

	 id := primitive.NewObjectID()
     announcement := classes.Announcement{
     	ID: id,
     	IDString: id.String(),
     	Title: title,
     	AuthorLogin: login,
     	Email: data["email"],
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
func AnnouncementInfo(c *gin.Context)  {
	fmt.Println("info")
	idData := struct {
		id string `json:"id"`
	}{}
    err := c.BindJSON(&idData)
    if err !=nil{
    	fmt.Println(err.Error())
    	c.String(400,err.Error())
		return
	}
    fmt.Println(idData.id)
    selector := make(map[string]interface{})
    selector["_idstr"]=idData.id
    data,err := session.ReadAnnouncements(selector)
    if err !=nil || data == nil || len(data) < 1{
    	fmt.Println(err.Error())
    	c.String(400,err.Error())
		return
	}else {
		fmt.Println(data[0])
		c.HTML(200,"anninfo.html",data[0])
		return
	}
}
func AddComment(c *gin.Context)  {
     var comment classes.Comment
     err := c.BindJSON(&comment)
     if err !=nil{
     	fmt.Println(err.Error())
     	c.String(400,err.Error())
		 return
	 }
	 err = session.Insert(comment,"coments")
     if err != nil{
		 fmt.Println(err.Error())
		 c.String(400,err.Error())
		 return
	 }
	 return
}

func AddToOrder(c *gin.Context)  {
	selector := make(map[string]interface{})
	var act classes.Activity
	login,err := c.Cookie("login")
	if err != nil{
		fmt.Println(err.Error())
		c.String(400,err.Error())
		return
	}
	selector["user_login"] = login
	err = c.BindJSON(&act)
	if err != nil{
		fmt.Println(err.Error())
		c.String(400,err.Error())
		return
	}
	order, err := session.ReadOrder(selector)
	if err != nil{
		fmt.Println(err.Error())
		c.String(400,err.Error())
		return
	}
	actList := order.ActivityList
	//actList = append(actList,act)
fmt.Println(actList)
}




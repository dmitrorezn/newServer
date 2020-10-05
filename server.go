package main

import (
	"ProjectMongoClient"
	"encoding/json"
	"fmt"
	classes "github.com/dmitrorezn/classes"
	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	//"unicode"
	"github.com/gin-gonic/contrib/static"
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
	router.LoadHTMLGlob("templ/*")
	router.Static("/auth/css","./templ")
	router.Static("/auth/script","./templ")
	router.Use(static.Serve("/photo", static.LocalFile("./photos/", true)))
	//router.Static("./scripts","C:/Users/dmitr/go/src/newServer/js_codes")
	router.Use(cors.Default())
	data := router.Group("/data")
	{
		data.Use(CheckUserTokenValidation)
		data.POST("/show", Show)
		data.GET("/userlogin",userlogin)
		data.GET("/announcements", Announcements)
		data.POST("/announcements", Announcements)
		data.GET("/activities", OrderActivities)
		//data.POST("/activities", OrderActivities)
		data.POST("/addToOrder", AddToOrder)
		data.POST("/delFromOrder", DeleteFromOrder)
		data.POST("/findAnnouncements", FindAnnouncements)

	}
	authordata := router.Group("/authordata")
	{
		authordata.Use(CheckAuthorTokenValidation)
		authordata.POST("/show", Show)
		authordata.GET("/userlogin",userlogin)
		authordata.GET("/announcements", Announcements)
		authordata.POST("/announcements", Announcements)
		authordata.POST("/delAnnouncement",DeleteAnnouncement)
		authordata.POST("/findAnnouncements", FindAnnouncements)

	}
	admindata := router.Group("/admindata")
	{
		admindata.Use(CheckAdminTokenValidation)
		admindata.POST("/show", Show)
		admindata.GET("/userlogin",userlogin)
		admindata.GET("/announcements", Announcements)
		admindata.POST("/announcements", Announcements)
		admindata.POST("/delAnnouncement",DeleteAnnouncement)
		admindata.POST("/findAnnouncements", FindAnnouncements)

	}
	auth := router.Group("/auth")
	{
		//auth.Use(CheckUserStatus)
		//auth.GET("/announcementinfo",AnnouncementInfo)
		//auth.POST("/announcementinfo",AnnouncementInfo)
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
		user.GET("/logout",Logout)
		user.GET("/anninfo",AnnInfoHtml)
		user.POST("/anninfo",AnnInfoHtml)
		user.GET("/account",Account)
		user.GET("/trash",Trash)
		user.GET("/announcements",Announcements)

	}
	author := router.Group("/author")
	{
		author.Use(CheckAuthorTokenValidation)
		author.GET("/page",AuthorPage)
		author.GET("/logout",Logout)
		author.POST("/addannouncement",AddAnnouncement)
		author.POST("/updateannouncement",UpdateAnnouncement)
		author.POST("/addImage",AddImage)
	}
	admin := router.Group("/admin")
	{
		admin.Use(CheckAdminTokenValidation)
		admin.GET("/page",AdminPage)
		admin.GET("/logout",Logout)
		admin.POST("/addannouncement",AddAnnouncement)
		admin.POST("/updateannouncement",UpdateAnnouncement)
		admin.POST("/addImage",AddImage)
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

func userlogin(c *gin.Context)  {
	login , err := c.Cookie("login")
	if err!= nil{
		c.String(400,err.Error())
		return
	}
	c.JSON(200,login)
}

func AnnInfoHtml(c *gin.Context)  {
	var selector map[string]interface{}
	curAnnId,err := c.Cookie("cur_ann_id")
	if err != nil {
		c.String(400,"Account->no cookie ",err.Error())
		return
	}
	c.SetCookie("cur_ann_id", "", -1, "/", "localhost", false, false)
	selector["_idstr"] = curAnnId
	announcements, err := session.ReadAnnouncements(selector)
	if err != nil || len(announcements)>1 {
		c.String(400,"ReadAnnouncements-> err ",err.Error())
	}
	fmt.Println(announcements[0])
	announcement := announcements[0]
	c.HTML(http.StatusOK,"anninfo.html",announcement)
	return
}

func AuthorPage(c *gin.Context)  {
	c.HTML(http.StatusOK, "author.html", nil)
}
func Account(c *gin.Context)  {
	var selector map[string]interface{}
	announcements,err := session.ReadAnnouncements(selector)
    if err != nil{
    	c.String(400,"ReadAnnouncements->"+err.Error())
		return
	}
	c.HTML(http.StatusOK, "index.html", announcements)
}

func AdminPage(c *gin.Context)  {
	c.HTML(http.StatusOK, "adminpage.html", nil)
}

func Show(c *gin.Context)  {
	var data interface{}
	err := c.BindJSON(&data)
	if err != nil{
		c.String(400,"BindJSON->"+err.Error())
		return
	}
	c.JSON(200, data)
}

//middleware token authorisation for user
func CheckUserTokenValidation(c *gin.Context) {
	_, err := c.Cookie("token")
	if err != nil{

		c.HTML(400, "signup.html", gin.H{
			"title": "authorisation",
			"error":err.Error(),
		})
		return
	}
	status,err := c.Cookie("status")
	if err != nil {
		fmt.Println(err.Error())
		c.HTML(400, "signup.html", gin.H{
			"title": "authorisation",
			"error":err.Error(),
		})
		return
	}
	if status != "user"{
		c.String(400,"wrong status")
		return
	}
	return
}

func CheckAuthorTokenValidation(c *gin.Context) {
	_, err := c.Cookie("token")
	if err != nil{
		c.HTML(200, "signup.html", gin.H{
			"title": "authorisation",
			"error":err.Error(),
		})
		return
	}
	status,err := c.Cookie("status")
	if err != nil {
		c.HTML(200, "signup.html", gin.H{
			"title": "authorisation",
			"error":err.Error(),
		})
		return
	}
	if status != "author"{
		c.String(400,"wrong status")
		return
	}
	return
}
func CheckAdminTokenValidation(c *gin.Context) {
	fmt.Println("hd")
	_, err := c.Cookie("token")
	if err != nil{

		c.HTML(400, "signup.html", gin.H{
			"title": "authorisation",
			"error":err.Error(),
		})
		return
	}
	status,err := c.Cookie("status")
	if err != nil {
		fmt.Println(err.Error())
		c.HTML(400, "signup.html", gin.H{
			"title": "authorisation",
			"error":err.Error(),
		})
		return
	}
	fmt.Println(status)
	if status != "admin"{
		c.String(400,"wrong status")
		return
	}
	return
}

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
				Account(context)

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
				return
			}
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
	var announcements  []classes.Announcement
	login, err := c.Cookie("login")
	if err !=  nil{
		fmt.Println(err.Error())
		c.String(400,err.Error())
		return
	}
	status, err := c.Cookie("status")
	if err !=  nil{
		fmt.Println(err.Error())
		c.String(400,err.Error())
		return
	}
	if status == "author" {
		selector["auth_login"] = login
		announcements, err = session.ReadAnnouncements(selector)
		if err != nil {
			fmt.Println(err.Error())
			c.String(400, err.Error())
			return
		}
	}else if status == "user"{
		announcements, err = session.ReadAnnouncements(selector)
		if err != nil {
			fmt.Println(err.Error())
			c.String(400, err.Error())
			return
		}
	}else if status == "admin"{
		announcements, err = session.ReadAnnouncements(selector)
		if err != nil {
			fmt.Println(err.Error())
			c.String(400, err.Error())
			return
		}
	}
	out,_ :=json.MarshalIndent(announcements,"  ","   ")
    fmt.Println(announcements[0].Title)
	fmt.Println(string(out))
	c.JSON(200,announcements)

}

func OrderActivities(c *gin.Context)  {
	var selector = make(map[string]interface{})
	login, err := c.Cookie("login")
	if err != nil{
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
	out,_ :=json.MarshalIndent(order.ActivityList,"  ","   ")
	fmt.Println(string(out))
	c.JSON(200,order.ActivityList)
	return
}

func UpdateAnnouncement(c *gin.Context)  {
	var selector = make(map[string]interface{})
	var set = make(map[string]interface{})
	var data map[string]string
	err := c.BindJSON(&data)
	if err != nil{
		fmt.Println(err.Error())
		return
	}
	price, err := strconv.ParseFloat(data["price"],64)
	if err!=nil{
		fmt.Println(err.Error())
		return
	}
	selector["_idstr"] = data["idstr"]
	set["title"] = data["title"]
	set["activity.name"] = data["name"]
	set["activity.type"] = data["type"]
	set["activity.price"] = price
	set["activity.description"] = data["description"]
	set["email"] = data["email"]
	set["phone_number"] = data["phone_number"]

	err = session.Update(selector,set,false,"announcements")
	if err != nil{
		fmt.Println(err.Error())
		c.String(400,err.Error())
		return
	}
	c.JSON(200,nil)
	return
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
	 startDates := strings.Split(data["start_dates"]," ")
	 fmt.Println(startDates)
	 id := primitive.NewObjectID()
	 actid := primitive.NewObjectID()
	 actidstr := strings.Split(actid.String(),"(")[1]
	 idstr := strings.Split(id.String(),"(")[1]
	 ids := []rune(idstr)
	 actids := []rune(actidstr)
	 ids = ids[1:len(ids)-2]
	 actids = actids[1:len(actids)-2]
	 imPath := string(ids)+"_image"
     announcement := classes.Announcement{
     	ID: id,
     	IDString: string(ids),
     	Title: title,
     	ImagePath: imPath,
     	AuthorLogin: login,
     	Email: data["email"],
     	StartWeekDays: startDates,
     	PhoneNumber: data["phone_number"],
     	Activity:classes.Activity{
     		ID: primitive.NewObjectID(),
     		IDSting: string(actids),
     		Name: data["name"],
     		Type: data["type"],
     		Price: price,
     		Description: data["description"],
		},
	 }
	 impath:= string(ids)+"_image"
	 c.SetCookie("annImPath",impath, 3600, "/", "localhost", false, false)
	 err = session.Insert(announcement,"announcements")
     if err != nil{
     	fmt.Println(err.Error())
     	c.String(400,err.Error())
     	return
	 }
	 c.JSON(200,nil)
     //fmt.Println(announcement)
}

func AddImage(c *gin.Context)  {
    imPath,err := c.Cookie("annImPath")
    if err != nil{
    	fmt.Println("path cookie"+err.Error())
    	c.String(400,"path cookie"+err.Error())
	}
	c.SetCookie("annImPath", "", -1, "/", "localhost", false, false)

	src, _, err := c.Request.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer src.Close()
	dst, err := os.Create("./photos/" + imPath)
	if err != nil {
		c.String(400,err.Error())
		return
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil{
		c.String(400, "io.Copy()->"+err.Error())
		return
	}
	c.JSON(200,nil)
	return
}


func AddToOrder(c *gin.Context)  {
	var dbannouns []classes.Announcement
	var dbact classes.Activity
	activities := make([]classes.Activity,2)
	info := make(map[string]string)
	orderselector := make(map[string]interface{})
	annselector := make(map[string]interface{})
	change := make(map[string]interface{})
	err1 :=c.BindJSON(&info)
	login,err2 := c.Cookie("login")
	if err1 !=nil || err2 !=nil{
		fmt.Println("no data id or cookie")
		c.String(400,"no data id or cookie")
		return
	}
    fmt.Println(info)
	annselector["_idstr"] = info["idstr"]
	dbannouns, err := session.ReadAnnouncements(annselector)
	dbact = dbannouns[0].Activity
	dbact.StartDate = info["day"]
	orderselector["user_login"] = login
	timems := strings.Split(time.Now().String(),".")[1]
	order, err := session.ReadOrder(orderselector)
	if err != nil {
		if err.Error() == "not found" {
			activities = append(activities, dbact)
			strid := dbact.IDSting
			dbact.IDSting = strid +"_"+ strings.Split(timems," ")[0]
			neworder := classes.Order{
				UserLogin:    login,
				ActivityList: activities,
				TotalPrice:   activities[0].Price,
			}
			err := session.Insert(neworder, "orders")
			if err != nil{
				fmt.Println(err.Error())
				c.String(400,err.Error())
				return
			}
			fmt.Println("inserted new order ")
			c.JSON(200,dbact)
			return
		}else{
		fmt.Println(err.Error())
		c.String(400,err.Error())
		return
		}
	}
	strid := dbact.IDSting
	dbact.IDSting = strid +"_"+ strings.Split(timems," ")[0]
	dbactivities := order.ActivityList
	dbactivities = append(dbactivities,dbact)
	change["activity_list"] = dbactivities
	err = session.Update(orderselector,change,false,"orders")
	if err != nil{
		fmt.Println("Update -> err ",err.Error())
		c.String(400,"Update -> err ",err.Error())
		return
	}
	fmt.Println("added activity")
	fmt.Println(dbact)
	acts := []classes.Activity{dbact}
	c.JSON(200,acts)
	return
}
func Logout(c *gin.Context) {
	// Clear the cookie
	c.SetCookie("token", "", -1, "", "", false, true)
	c.SetCookie("status", "", -1, "", "", false, true)
	c.SetCookie("login", "", -1, "", "", false, true)
	c.Set("isLoggedIn", false)
	// Redirect to the home page
	c.Redirect(http.StatusFound, "/auth/signform")
}
func DeleteFromOrder(c *gin.Context) {

	selector := make(map[string]interface{})
	set := make(map[string]interface{})
	login, err := c.Cookie("login")
	if err != nil{
		fmt.Println(err.Error())
		c.String(400,err.Error())
		return
	}

	selector["user_login"] = login
	order, err := session.ReadOrder(selector)
	actList := order.ActivityList
	var delId string
	err = c.BindJSON(&delId)
	if err != nil {
		fmt.Println("deleteFromOrder() ->", err.Error())
		return
	}
	for k := range actList {
		if actList[k].IDSting == delId {
			actList[k] = actList[len(actList)-1]
			actList = actList[:len(actList)-1]
			break
		}
	}
	fmt.Println("delete Activity  ->",  delId)
	set["activity_list"] = actList
	err = session.Update(selector,set,false,"orders")
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Removed Activity :",delId )
	}
	c.JSON(200,nil)
}

func DeleteAnnouncement(c *gin.Context)  {
	selector := make(map[string]interface{})
	var idstr string
	err := c.BindJSON(&idstr)
	if err != nil {
		c.String(400,err.Error())
		return
	}
	selector["_idstr"] = idstr
	err = session.Delete(selector,false,"announcements")
	if err != nil{
		c.String(400,"Delete->" + err.Error())
		return
	}
	c.JSON(200,nil)
	return
}

func FindAnnouncements(c *gin.Context)  {
     titleselector := make(map[string]interface{})
	 nameselector := make(map[string]interface{})
	 typeselector := make(map[string]interface{})
	 var input string
     var announcements []classes.Announcement
     err :=  c.BindJSON(&input)
     if err != nil{
     	c.String(400,err.Error())
		 return
	 }
	 titleselector["title"] = input
	 announcements, err = session.ReadAnnouncements(titleselector)
	 if err != nil{
		 nameselector["activity.name"] = input
		 announcements, err = session.ReadAnnouncements(nameselector)
		 if err != nil{
			 typeselector["activity.type"] = input
			 announcements, err = session.ReadAnnouncements(typeselector)
             if err != nil{
             	c.JSON(400,nil)
             	return
			 }
		 }
	 }
     c.JSON(200,announcements)
	 return
}
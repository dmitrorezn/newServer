package main

import (
	"ProjectMongoClient"
	"fmt"
	"github.com/gin-gonic/gin"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
)

var session *ProjectMongoClient.DBSession

func main() {
	var err error
	tablesMap := make(map[string]string)
	tablesMap["users"] ="userstable"
	tablesMap["activities"] = "activitiestable"
	tablesMap["orders"] = "orderstable"
	tablesMap["announcements"] = "announcementstable"

	session, err = ProjectMongoClient.NewSession("toursdb",tablesMap)
	if err != nil{
		fmt.Println(err.Error())
		return
	}
	fmt.Println(session)
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("./style","./templates")
	RouterGroupsInit(router)
}


func RouterGroupsInit(router *gin.Engine){

	auth := router.Group("/auth")
	{
		auth.Use(CheckUserStatus)
		auth.GET("/regform",RegForm)
		auth.GET("/sigform",SignForm)


	}

	user := router.Group("/user")
	{
		user.Use(CheckUserTokenValidation)
		user.GET("/account",Account)
	}

	admin := router.Group("/admin")
	{
		admin.GET("/page",AdminPage)
	}

}

func RegForm(c *gin.Context)  {
	c.HTML(http.StatusOK, "registration.html", nil)
}
func SignForm(c *gin.Context)  {
	c.HTML(http.StatusOK, "signup.html", nil)
}
func Account(c *gin.Context)  {
	c.HTML(http.StatusOK, "account.html", nil)
}
func AdminPage(c *gin.Context)  {
	c.HTML(http.StatusOK, "adminpage.html", nil)
}

//middleware token authorisation for user
func CheckUserTokenValidation(c *gin.Context) {
	_, err := c.Cookie("token")
	if err != nil {
		c.HTML(200, "signin.html", gin.H{
			"title": "authorisation",
		})
		return
	}
	return
}

//middleware token authorisation for user
func CheckUserStatus(c *gin.Context) {
	status, ok := c.GetPostForm("status")
	if !ok {
		c.HTML(200, "signin.html", gin.H{
			"title": "authorisation",
			"status":status,
		})
		return
	}
	return
}

func Registration(context *gin.Context) {
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
					user, err := session.CheckUserInDB(login, email, password)
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

		if login != "" && password != "" && email != "" {
			user, err := session.CheckUserInDB(login, email, password)
			if err != nil {
				fmt.Println(err.Error())
				context.String(http.StatusNoContent, "user already exists ", err.Error())
				return
			}
			//fmt.Println("user:", user)
			err = session.Insert(user,"users")
			if err != nil {
				fmt.Println(err.Error())
				context.String(http.StatusInternalServerError, "err user add ", err.Error())
				return
			}
			fmt.Println("Sucsesfully added User ", login, "id ", user.ID)
			context.String(http.StatusOK, "Sucsesfully added User %s", login)
		} else {
			fmt.Println("no password or login or email info")
		}
	}
	return
}

func SignUp(context *gin.Context) {
	if context.Request.Header.Get("content-type") == "application/json" {
		mymap := make(map[string]string)
		err := context.BindJSON(&mymap)
		if err != nil {
			fmt.Println(err.Error())
			context.String(http.StatusOK, "error 2 ")
			return
		}
		if login, ok := mymap["login"]; ok {
			if password, ok := mymap["password"]; ok {
				user, err := session.CheckUserPassword(login, password)
				if err != nil {
					fmt.Println(err.Error())
					context.String(http.StatusOK, err.Error())
					return
				}
				//sending token as a cookie
				tokenstring := user.GetToken()
				context.SetCookie("token", tokenstring, 5*60, "/", "localhost", false, false)
				//return HTML
				context.HTML(http.StatusOK, "account.html", gin.H{
					"Username": user.Username,
					"Email":    user.Email,
				})
				return

			}
		}
		context.String(http.StatusOK, "sth wrong")
		return
	} else {

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
			//return HTML
			//context.String(http.StatusOK,"user:",user)
			//resp, err := http.Get(fmt.Sprintf("localhost:8080/user/account?token=%s&username=%s&email=%s", tokenstring, user.Username,user.Email))
			//if err != nil {
			//	fmt.Println(err.Error())
			//	context.String(http.StatusConflict, err.Error())
			//}
			//fmt.Println(resp)
			context.HTML(http.StatusOK, "account.html", gin.H{
				"Username": user.Username,
				"Email":    user.Email,
				//"Id": user.IDString,
			})
			return
		} else {
			fmt.Println("no password or email")
			context.String(http.StatusBadRequest, "no password or email")
			return
		}
	}
}



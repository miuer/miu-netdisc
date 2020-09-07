package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/miuer/miu-netdisc/filestore/model/mysql"
	"github.com/miuer/miu-netdisc/filestore/utils"
)

const (
	pwdSalt = "miuer"
)

// RegisterHandler -
func (ctl *Controller) RegisterHandler(w http.ResponseWriter, r *http.Request) {

	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	confirm := r.PostFormValue("confirm")
	email := r.PostFormValue("email")
	phone := r.PostFormValue("phone")

	if password != confirm {
		w.WriteHeader(http.StatusNotAcceptable)
		io.WriteString(w, errors.New("ReSetPasswordNotMatch").Error())
		return
	}

	if !utils.CheckUsernameValidity(username) {
		w.WriteHeader(http.StatusNotAcceptable)
		io.WriteString(w, errors.New("UsernameFormatNotMatch").Error())
		return
	}

	if !utils.CheckEmailValidity(email) {
		w.WriteHeader(http.StatusNotAcceptable)
		io.WriteString(w, errors.New("EmailFormatNotMatch").Error())
		return
	}

	if !utils.CheckPhoneValidity(phone) {
		w.WriteHeader(http.StatusNotAcceptable)
		io.WriteString(w, errors.New("PhoneFormatNotMatch").Error())
		return
	}

	sha1Pwd := utils.Sha1Byte([]byte(password + pwdSalt))

	user := &mysql.User{
		Username: username,
		Password: sha1Pwd,
		Email:    email,
		Phone:    phone,
	}

	err := mysql.AddNewUser(ctl.Writer, user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to register new account, err:"+err.Error())
		return
	}

	http.Redirect(w, r, "/user/registerSucceed", http.StatusFound)
}

// RegisterSucceedHandler -
func (ctl *Controller) RegisterSucceedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("register Successfully"))
}

// CheckUserNameHandler -
func (ctl *Controller) CheckUserNameHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")

	if !utils.CheckUsernameValidity(username) {
		w.WriteHeader(http.StatusNotAcceptable)
		io.WriteString(w, errors.New("UsernameFormatNotMatch").Error())
		return
	}

	id, _, err := mysql.GetIDAndPwdByUsername(ctl.Reader, username)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Failed to get id by username, err:"+err.Error())
		return
	}

	if id > 0 {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "the username has been registered")
	} else {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "the username can be used")
	}
}

// CheckUserEmailHandler -
func (ctl *Controller) CheckUserEmailHandler(w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")

	if !utils.CheckEmailValidity(email) {
		w.WriteHeader(http.StatusNotAcceptable)
		io.WriteString(w, errors.New("EmailFormatNotMatch").Error())
		return
	}

	id, err := mysql.GetIDByEmail(ctl.Reader, email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to get id by email, err:"+err.Error())
		return
	}

	if id > 0 {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "the email has been registered")
	} else {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "the email can be used")
	}
}

// CheckUserPhoneHandler -
func (ctl *Controller) CheckUserPhoneHandler(w http.ResponseWriter, r *http.Request) {
	phone := r.PostFormValue("phone")

	if !utils.CheckPhoneValidity(phone) {
		w.WriteHeader(http.StatusNotAcceptable)
		io.WriteString(w, errors.New("PhoneFormatNotMatch").Error())
		return
	}

	id, err := mysql.GetIDByPhone(ctl.Reader, phone)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to get id by phone, err:"+err.Error())
		return
	}

	if id > 0 {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "the phone has been registered")
	} else {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "the phone can be used")
	}
}

// LoginHandler -
func (ctl *Controller) LoginHandler(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	userID, sha1Pwd, err := mysql.GetIDAndPwdByUsername(ctl.Reader, username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to get user info, err:"+err.Error())
		return
	}

	if userID <= 0 {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Invalid username")
		return
	}

	if !(sha1Pwd == utils.Sha1Byte([]byte(password+pwdSalt))) {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, "Incorrect password")
		return
	}

	token := utils.GenerateToken(username)
	cookie := http.Cookie{
		Name:     "tk",
		Path:     "/",
		Value:    token,
		HttpOnly: true,
	}

	mysql.ReplaceToken(ctl.Writer, userID, token)
	http.SetCookie(w, &cookie)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login succeed"))
}

package server

import (
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
)

const (
	signInView    = "sign_in.view.html"
	signUpView    = "sign_up.view.html"
	dashboardView = "dashboard.view.html"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	if userSignedIn(r) {
		http.Redirect(w, r, dashboard, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, signIn, http.StatusSeeOther)
	}
}

func getDashboard(w http.ResponseWriter, r *http.Request) {
	if userSignedIn(r) {
		app.Data.CSRFToken = nosurf.Token(r)
		if err := writeView(w, dashboardView); err != nil {
			fmt.Println(err)
		}
	} else {
		http.Redirect(w, r, signIn, http.StatusSeeOther)
	}
}

func getSignIn(w http.ResponseWriter, r *http.Request) {
	if userSignedIn(r) {
		http.Redirect(w, r, dashboard, http.StatusSeeOther)
	} else {
		app.Data.CSRFToken = nosurf.Token(r)
		if err := writeView(w, signInView); err != nil {
			fmt.Println(err)
		}
	}
}

func postSignIn(w http.ResponseWriter, r *http.Request) {
	params := []string{"username", "password"}
	err := validateFormParams(r, params)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, signIn, http.StatusSeeOther)
		return
	}

	user, err := app.database.SelectUserByUsername(r)
	app.Data.CurrentUser = user
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, signIn, http.StatusSeeOther)
		return
	}

	err = validatePassword(r.PostFormValue("password"), user.EncryptedPassword)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, signIn, http.StatusSeeOther)
		return
	}
	user.EncryptedPassword = ""

	err = app.database.UpdateUserLastLogin(user)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, signIn, http.StatusSeeOther)
		return
	}

	_ = app.session.RenewToken(r.Context())
	app.session.Put(r.Context(), "signedIn", true)
	http.Redirect(w, r, dashboard, http.StatusSeeOther)
}

func getSignUp(w http.ResponseWriter, r *http.Request) {
	if userSignedIn(r) {
		http.Redirect(w, r, dashboard, http.StatusSeeOther)
	} else {
		app.Data.CSRFToken = nosurf.Token(r)
		if err := writeView(w, signUpView); err != nil {
			fmt.Println(err)
		}
	}
}

func postSignUp(w http.ResponseWriter, r *http.Request) {
	err := validateSignUpForm(r)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, signUp, http.StatusSeeOther)
		return
	}

	ic, err := app.database.SelectInvitationCode(r.PostFormValue("code"))
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, signUp, http.StatusSeeOther)
		return
	}

	err = validateDate(ic.Validity)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, signUp, http.StatusSeeOther)
		return
	}

	ep, err := encryptPassword(r.PostFormValue("password"))
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, signUp, http.StatusSeeOther)
		return
	}

	err = app.database.InsertUser(r, ep)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, signUp, http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, signIn, http.StatusSeeOther)
}

func postSignOut(w http.ResponseWriter, r *http.Request) {
	app.Data.CurrentUser = nil
	_ = app.session.Destroy(r.Context())
	_ = app.session.RenewToken(r.Context())
	http.Redirect(w, r, signIn, http.StatusSeeOther)
}

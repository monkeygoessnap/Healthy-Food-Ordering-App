package server

import (
	"FPproject/Frontend/log"
	"FPproject/Frontend/models"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
)

var tpl *template.Template

func Run() {
	tpl = template.Must(template.ParseGlob("templates/*"))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/favicon.ico", http.NotFoundHandler())

	http.HandleFunc("/healthcheck", healthCheck)
	http.HandleFunc("/", index)
	http.HandleFunc("/home", home)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/register", register)

	http.HandleFunc("/profile", profile)
	http.HandleFunc("/editprofile", editprofile)
	http.HandleFunc("/browse", browse)
	http.HandleFunc("/browse/res", res)
	http.HandleFunc("/cart", cart)

	log.Info.Println("Frontend running at :8181")
	log.Error.Println(http.ListenAndServe(":8181", nil))
}

func cart(w http.ResponseWriter, r *http.Request) {

	var tpldata []interface{}
	var cart []models.CartItem
	data, status := newRequest(r, http.MethodGet, "/allci", nil)
	if status != 200 {
		tpl.ExecuteTemplate(w, "err.html", nil)
		return
	}
	json.Unmarshal(data, &cart)
	var foods []models.Food
	for _, v := range cart {
		var food models.Food
		fooddata, _ := newRequest(r, http.MethodGet, "/food/"+v.ID, nil)
		json.Unmarshal(fooddata, &food)
		foods = append(foods, food)
	}
	var uh models.UserHealth
	data, _ = newRequest(r, http.MethodGet, "/uh", nil)
	json.Unmarshal(data, &uh)

	calData := tCal(cart, foods, uh)
	tpldata = append(tpldata, cart, foods, calData)
	tpl.ExecuteTemplate(w, "cart.html", tpldata)
}

type Tcal struct {
	Cal    int
	UCal   int
	Target string
	Msg    string
	Color  string
}

func tCal(carts []models.CartItem, foods []models.Food, uh models.UserHealth) Tcal {
	var cl Tcal
	for i, v := range foods {
		cl.Cal = (v.Calories * carts[i].Qty) + cl.Cal
	}
	userCal := calories(uh.Gender, uh.DOB, uh.Active, uh.Height, uh.Weight)
	switch uh.Target {
	case "lose":
		if cl.Cal > userCal {
			cl.Msg = "Calories exceeded!"
			cl.Color = "red"
		} else {
			cl.Msg = "Within calories goal"
			cl.Color = "green"
		}
	case "gain":
		if cl.Cal > userCal {
			cl.Msg = "Calories goal achieved!"
			cl.Color = "green"
		} else {
			cl.Msg = "Calories goal not achieved yet"
			cl.Color = "yellow"
		}
	case "maintain":
		if cl.Cal > int((float32(userCal) * 1.05)) {
			cl.Msg = "Calories exceeded!"
			cl.Color = "red"
		} else if cl.Cal < int((float32(userCal) * 1.05)) {
			cl.Msg = "Calories goal not achieved yet"
			cl.Color = "yellow"
		} else {
			cl.Msg = "Calories goal achieved!"
			cl.Color = "green"
		}
	}
	cl.UCal = userCal
	cl.Target = uh.Target
	return cl
}

func res(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	var tpldata []interface{}
	name := r.URL.Query().Get("name")
	var add models.Address
	data, _ := newRequest(r, http.MethodGet, "/mercadd/"+id, nil)
	json.Unmarshal(data, &add)
	var foods []models.Food
	data, _ = newRequest(r, http.MethodGet, "/allfood/"+id, nil)
	json.Unmarshal(data, &foods)
	tpldata = append(tpldata, name, add, foods)

	// var ci []models.CartItem
	// data, _ = newRequest(r, http.MethodGet, "/allci", nil)

	if r.Method == http.MethodPost {
		id := r.FormValue("add")
		qty, _ := strconv.Atoi(r.FormValue(id))
		new := models.CartItem{
			ID:  id,
			Qty: qty,
		}
		_, status := newRequest(r, http.MethodPost, "/ci", new)
		if status != 200 {
			tpl.ExecuteTemplate(w, "err.html", nil)
			return
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "res.html", tpldata)
}

func browse(w http.ResponseWriter, r *http.Request) {

	var mercs []models.User
	data, _ := newRequest(r, http.MethodGet, "/merc", nil)
	json.Unmarshal(data, &mercs)
	if r.Method == http.MethodPost {
		id := r.FormValue("id")
		name := r.FormValue("name")
		http.Redirect(w, r, "/browse/res?id="+id+"&name="+name, http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "browse.html", mercs)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {

	_, status := newRequest(r, http.MethodGet, "/healthcheck", nil)

	log.Info.Println("Healthcheck code: %v", status)
	tpl.ExecuteTemplate(w, "healthcheck.html", status)
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "index.html", nil)
}

func home(w http.ResponseWriter, r *http.Request) {

	var tpldata []interface{}
	data, status := newRequest(r, http.MethodGet, "/user", nil)
	var user models.User
	json.Unmarshal(data, &user)
	data, _ = newRequest(r, http.MethodGet, "/uh", nil)
	var uh models.UserHealth
	json.Unmarshal(data, &uh)
	cal := models.AddData{
		Calories: calories(uh.Gender, uh.DOB, uh.Active, uh.Height, uh.Weight),
		Age:      ageCal(uh.DOB),
		BMI:      bmi(uh.Weight, uh.Height),
	}
	tpldata = append(tpldata, user, uh, cal)
	if status > 200 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "home.html", tpldata)
}

func login(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		jsonData := map[string]string{
			"username": r.FormValue("username"),
			"password": r.FormValue("password"),
		}
		data, status := newRequest(r, http.MethodPost, "/login", jsonData)
		tpldata := map[string]string{}
		json.Unmarshal(data, &tpldata)
		if status != 200 {
			tpl.ExecuteTemplate(w, "login.html", tpldata)
			return
		}
		setCookie(w, r, tpldata["access_token"], tpldata["expire"])
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(w, "login.html", nil)
}

func register(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		//TODO data sanitization
		jsonData := models.User{
			Username: r.FormValue("username"),
			Name:     r.FormValue("name"),
			UserType: r.FormValue("usertype"),
			Password: hash(r.FormValue("password")),
		}
		data, _ := newRequest(r, http.MethodPost, "/register", jsonData)
		tpldata := map[string]string{}
		json.Unmarshal(data, &tpldata)

		tpl.ExecuteTemplate(w, "register.html", tpldata)
		return
	}
	tpl.ExecuteTemplate(w, "register.html", nil)
}

func profile(w http.ResponseWriter, r *http.Request) {

	var tpldata []interface{}
	userdata, status := newRequest(r, http.MethodGet, "/user", nil)
	if status != 200 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var user models.User
	json.Unmarshal(userdata, &user)
	adddata, _ := newRequest(r, http.MethodGet, "/add", nil)
	var add models.Address
	json.Unmarshal(adddata, &add)
	healthdata, _ := newRequest(r, http.MethodGet, "/uh", nil)
	var health models.UserHealth
	json.Unmarshal(healthdata, &health)
	tpldata = append(tpldata, user, add, health)
	tpl.ExecuteTemplate(w, "profile.html", tpldata)
}

func editprofile(w http.ResponseWriter, r *http.Request) {

	var tpldata []interface{}
	userdata, status := newRequest(r, http.MethodGet, "/user", nil)
	if status != 200 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var user models.User
	json.Unmarshal(userdata, &user)
	adddata, _ := newRequest(r, http.MethodGet, "/add", nil)
	var add models.Address
	json.Unmarshal(adddata, &add)
	if add.ID == "" {
		newRequest(r, http.MethodPost, "/add", nil)
	}
	healthdata, _ := newRequest(r, http.MethodGet, "/uh", nil)
	var health models.UserHealth
	json.Unmarshal(healthdata, &health)
	if health.ID == "" {
		newRequest(r, http.MethodPost, "/uh", nil)
	}
	tpldata = append(tpldata, user, add, health)

	if r.Method == http.MethodPost {

		userJson := models.User{
			Name: r.FormValue("name"),
		}
		newRequest(r, http.MethodPut, "/user", userJson)

		addJson := models.Address{
			Postal: r.FormValue("postal"),
			Floor:  r.FormValue("floor"),
			Unit:   r.FormValue("unit"),
		}
		newRequest(r, http.MethodPut, "/add", addJson)

		healthJson := models.UserHealth{
			Gender: r.FormValue("gender"),
			Height: convFloat(r.FormValue("height")),
			Weight: convFloat(r.FormValue("weight")),
			DOB:    r.FormValue("dob"),
			Active: r.FormValue("active"),
			Target: r.FormValue("target"),
		}
		newRequest(r, http.MethodPut, "/uh", healthJson)

		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "editprofile.html", tpldata)
}

func logout(w http.ResponseWriter, r *http.Request) {
	token := &http.Cookie{
		Name:   "token",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
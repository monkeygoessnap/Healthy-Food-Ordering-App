package server

import (
	"FPproject/Frontend/log"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func getToken(r *http.Request) string {
	token, err := r.Cookie("token")
	if err != nil {
		return ""
	}
	return token.Value
}

func newRequest(req *http.Request, method, url string, jsonD interface{}) ([]byte, int) {
	base := "http://localhost:8080/api/v1"
	jsonV, _ := json.Marshal(jsonD)
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
	}
	r, _ := http.NewRequest(method, base+url, bytes.NewBuffer(jsonV))
	r.Header.Set("Content-type", "application/json")
	r.Header.Set("access_token", getToken(req))
	resp, err := client.Do(r)
	if err != nil {
		log.Warning.Println(err)
		return nil, http.StatusServiceUnavailable
	}
	data, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return data, resp.StatusCode
}

func convFloat(input string) float32 {
	output, err := strconv.ParseFloat(input, 32)
	if err != nil {
		log.Warning.Println(err)
		return 0
	}
	return float32(output)
}

func hash(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hash)
}

func setCookie(w http.ResponseWriter, r *http.Request, tk string, expiry string) {

	unix, _ := strconv.Atoi(expiry)
	expiryUnix := time.Unix(int64(unix), 0)

	token := &http.Cookie{
		Name:     "token",
		Value:    tk,
		HttpOnly: true,
		Expires:  expiryUnix,
		Path:     "/",
		Secure:   true,
	}
	http.SetCookie(w, token)
}

func bmi(weight float32, height float32) float32 {
	/* According to HealthHub,
	BMI value of 23 and above indicates that yout weight is outside of the healthy weight range for
	your height
	*/
	return weight / height / height * 10000

}

func ageCal(dob string) int {

	layout := "020106" //ddmmyy
	t1 := time.Now()
	t2, err := time.Parse(layout, dob)
	if err != nil {
		log.Warning.Println(err)
		return 0
	}
	diff := int(t1.Sub(t2).Hours() / 24 / 365)
	return diff
}

func calories(gender, dob, active string, height, weight float32) int {
	//https://www.k-state.edu/paccats/Contents/PA/PDF/Physical%20Activity%20and%20Controlling%20Weight.pdf
	/*activity multiplier
	low = 1.2 x BMR (little to no exercise)
	moderate = 1.55 x BMR (moderate exercise 6 times a week)
	high = 1.9 x BMR (training 2 or more times a day)
	*/
	//Using Harris-Benedict Formula
	age := float32(ageCal(dob))
	var mul float32
	switch active {
	case "low":
		mul = 1.2
	case "moderate":
		mul = 1.55
	case "high":
		mul = 1.9
	default:
		mul = 1.0
	}
	switch gender {
	case "male":
		bmr := 66 + (13.7 * weight) + (5 * height) - (6.8 * age)
		return int(bmr * mul)
	case "female":
		bmr := 655 + (9.6 * weight) + (1.8 * height) - (4.7 * age)
		return int(bmr * mul)
	default:
		return 0
	}
	return 0
}
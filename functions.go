package main

import "fmt"

const base62Digits = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// create new short url
func (app *application) create_shortner(id int64) string {
	url := app.base62Encode(id)
	return fmt.Sprintf("smile%s", url)

}

// create new short url
func (app *application) create_custom_shortner(url string) string {
	return fmt.Sprintf("smile%s", url)

}

// get short url
func (app *application) get_shortner(url string) string {
	return fmt.Sprintf("smile%s", url)

}

func (app *application) base62Encode(id int64) string {
	url := ""
	i := id
	for i > 0 {
		remanider := i % 62
		url = string(base62Digits[remanider]) + url
		i = i / 62
	}

	return "xyz" + url
}

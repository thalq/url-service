package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

func main() {
	endpoint := "http://localhost:8080/"
	data := url.Values{}
	fmt.Println("Введите длинный URL")
	reader := bufio.NewReader(os.Stdin)
	long, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	long = strings.TrimSuffix(long, "\n")
	data.Set("url", long)
	client := resty.New()

	resp, err := client.R().
		SetBody(data.Encode()).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post(endpoint)
	if err != nil {
		panic(err)
	}

	fmt.Println("Статус-код ", resp.StatusCode())
	bodyString := resp.String()
	fmt.Println("Response Body as string:", bodyString)

}

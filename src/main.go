package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"v1/mapper"
)

func main() {
	daysOfRetentionString := os.Getenv("INPUT_RETENTION-DAYS")
	projectName := os.Getenv("INPUT_PROJECT-NAME")
	ownerAccount := os.Getenv("INPUT_OWNER-ACCOUNT")
	tokenAccess := os.Getenv("INPUT_ACCESS-TOKEN")
	baseUrlGithub := "https://api.github.com/repos"

	if strings.TrimSpace(daysOfRetentionString) == "" || strings.TrimSpace(projectName) == "" || strings.TrimSpace(ownerAccount) == "" || strings.TrimSpace(tokenAccess) == "" {
		fmt.Println("Some variable are empty")
		os.Exit(1)
	}

	fmt.Println("Starting the code ...")

	urlCompleta := fmt.Sprintf("%s/%s/%s/actions/artifacts", baseUrlGithub, ownerAccount, projectName)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", urlCompleta, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenAccess))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)

	fmt.Println("Getting the artifacts ...")

	if err != nil {
		log.Fatalf("Error on get your request: %v\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Error on read the body: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Mapping the return ...")

	var bodyResponse = mapper.Artifacts{}
	err = json.Unmarshal(bodyBytes, &bodyResponse)
	if err != nil {
		log.Fatalf("Failed in parser the json: %v\n", err)
		os.Exit(1)
	}

	dayOfToday := time.Now()
	daysOfRetention, err := strconv.Atoi(daysOfRetentionString)

	if err != nil {
		log.Fatalf("Failed to convert the string to int: %v\n", err)
		os.Exit(1)
	}

	if bodyResponse.TotalCount == 0 {
		fmt.Println("There aren't any artifact to delete on project", projectName)
		os.Exit(0)
	}

	dataToRetention := dayOfToday.AddDate(0, 0, -1*daysOfRetention)
	fmt.Println("Total artifacts found:", bodyResponse.TotalCount)
	itemsDeleted := 0

	for _, artifact := range bodyResponse.Artifacts {
		if artifact.CreatedAt.After(dataToRetention) {
			log.Printf("The artifact \"%d\" aren't deleted because of date are great than date of retention ...\n", artifact.Id)
			continue
		}

		client := &http.Client{}
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/%d", urlCompleta, artifact.Id), nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenAccess))
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		defer res.Body.Close()
		itemsDeleted += 1
		log.Printf("Artifact with id \"%d\" removed with success ...\n", artifact.Id)
	}
	fmt.Println("Total items deleted:", itemsDeleted)
}

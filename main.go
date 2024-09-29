package main

import (
	"bufio"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hackyeah-aezakmi/safeshop/ai"
	"github.com/likexian/whois"
	whoisparser "github.com/likexian/whois-parser"
	"github.com/sashabaranov/go-openai"
)

func main() {
	log.Println("Starting SafeShop server")

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	domainsChan := periodicFetchMaliciousDomains(1 * time.Hour)
	var latestDomains []string

	go func() {
		for domains := range domainsChan {
			latestDomains = domains
		}
	}()

	r := gin.Default()

	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.POST("/check/:hostname", func(c *gin.Context) {
		hostname := c.Param("hostname")

		decodedHostname, err := base64.StdEncoding.DecodeString(hostname)
		if err != nil {
			log.Println("Error decoding domain:", err)
			c.String(http.StatusBadRequest, "Invalid domain encoding")
			return
		}
		hostname = string(decodedHostname)
		log.Println("Checking domain:", hostname)

		// Use latestDomains instead of receiving from the channel
		for _, maliciousDomain := range latestDomains {
			if strings.HasSuffix(hostname, maliciousDomain) {
				c.String(http.StatusOK, "Received request for hostname: %s", hostname)
				return
			}
		}

		// Check domain age
		domainAge, err := getDomainAgeInDays(hostname)
		if err != nil {
			log.Printf("Error getting domain age: %v", err)
			c.String(http.StatusInternalServerError, "Error getting domain age")
			return
		}

		if domainAge < 60 {
			log.Printf("Domain %s is less than 60 days old", hostname)
		} else if domainAge < 365 {
			log.Printf("Domain %s is between 60 and 365 days old", hostname)
		} else {
			log.Printf("Domain %s is more than 1 year old", hostname)
		}

		// Get AI domain score
		score, err := ai.GetDomainScore(*client, hostname)
		if err != nil {
			log.Printf("Error getting domain score: %v", err)
			c.String(http.StatusInternalServerError, "Error getting domain score")
			return
		}
		log.Printf("Domain %s score: %d", hostname, score)

		c.String(http.StatusOK, "Received request for hostname: %s", hostname)
	})

	r.Run(":8080")
}

func getDomainAgeInDays(domain string) (int, error) {
	whoisData, err := whois.Whois(domain)
	if err != nil {
		return 0, err
	}

	whoisParser, err := whoisparser.Parse(whoisData)
	if err != nil {
		return 0, err
	}

	createdDate, err := time.Parse(time.RFC3339, whoisParser.Domain.CreatedDate)
	if err != nil {
		return 0, err
	}

	daysSinceCreation := time.Since(createdDate).Hours() / 24

	return int(daysSinceCreation), nil
}

func fetchMaliciousDomains() ([]string, error) {
	resp, err := http.Get("https://hole.cert.pl/domains/v2/domains.txt")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var domains []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		domain := strings.TrimSpace(scanner.Text())
		if domain != "" {
			domains = append(domains, domain)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	log.Printf("Fetching malicious domains completed, found %d domains", len(domains))

	return domains, nil
}

func periodicFetchMaliciousDomains(interval time.Duration) <-chan []string {
	domainsChan := make(chan []string)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in periodicFetchMaliciousDomains: %v", r)
			}
		}()

		log.Println("Starting periodic fetch of malicious domains")

		// Fetch immediately on start
		if domains, err := fetchMaliciousDomains(); err == nil {
			domainsChan <- domains
		} else {
			log.Printf("Error fetching malicious domains: %v", err)
		}

		// Set up ticker for periodic fetches
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			domains, err := fetchMaliciousDomains()
			if err != nil {
				log.Printf("Error fetching malicious domains: %v", err)
			} else {
				domainsChan <- domains
			}
		}
	}()

	return domainsChan
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

type cardNumberRequest struct {
	CardNumber string `json:"card_number"`
}

type cardNumberResponse struct {
	cardNumberRequest
	Valid   bool   `json:"valid"`
	Network string `json:"network,omitempty"`
}

func isValidCardNumber(cardNumber string) bool {
	var sum int
	parity := (len(cardNumber) - 1) % 2
	digits := []rune(cardNumber)
	checksumDigit, _ := strconv.Atoi(string(digits[len(digits)-1]))
	for i := len(digits) - 2; i >= 0; i-- {
		myDigit, _ := strconv.Atoi(string(digits[i]))

		if i%2 == parity {
			sum += myDigit
		} else if myDigit > 4 {
			sum += (2 * myDigit) - 9
		} else {
			sum += 2 * myDigit
		}
	}

	return checksumDigit == 10-(sum%10)
}

func getRange(ranges string) (int, int) {
	tab := strings.Split(ranges, "_")
	inf, _ := strconv.Atoi(tab[0])
	sup, _ := strconv.Atoi(tab[1])

	return inf, sup
}

func getNetWorkName(cardNumber string) string {

	issuingNetwork := map[string]string{
		"34": "American Express", "37": "American Express",
		"5610": "Bankcard", "560221_560225": "Bankcard",
		"31":   "China T-Union",
		"62":   "China UnionPay",
		"36":   "Diners Club International",
		"55":   "Diners Club United States & Canada",
		"6011": "Discover Card", "644_649": "Discover Card",
		"647": "Discover Card", "622126_622925": "Discover Card",
		"60400100_60420099": "UkrCard",
		"60":                "RuPay", "81": "RuPay", "82": "RuPay", "508": "RuPay", "353": "RuPay", "356": "RuPay",
		"636":       "InterPayment",
		"637_639":   "InstaPayment",
		"3528_3589": "JCB",
		"676770":    "Maestro UK", "676774": "Maestro UK",
		"5018": "Maestro", "5020": "Maestro", "5038": "Maestro", "5893": "Maestro", "6304": "Maestro", "6761": "Maestro", "6762": "Maestro",
		"6763": "Maestro",
		"5019": "Dankort", "4571": "Dankort",
		"2200_2204": "Mir",
		"2205":      "BORICA",
		"2221_2720": "Mastercard", "51_55": "Mastercard",
		"4903": "Switch", "4905": "Switch", "4911": "Switch", "4936": "Switch", "564182": "Switch", "633110": "Switch", "6333": "Switch", "6759": "Switch | Maestro",
		"65": "Troy | Discover Card", "9792": "Troy",
		"4026": "Visa Electron", "417500": "Visa Electron", "4508": "Visa Electron", "4844": "Visa Electron", "4913": "Visa Electron", "4917": "Visa Electron",
		"1":             "UATP",
		"506099_506198": "Verve", "650002_650027": "Verve", "507865_507964": "Verve",
		"357111": "LankaPay",
		"9704":   "Napas",
	}
	for iin, netw := range issuingNetwork {
		if strings.Contains(iin, "_") {
			inf, sup := getRange(iin)
			for i := inf; i <= sup; i++ {
				if strings.HasPrefix(cardNumber, strconv.Itoa(i)) {
					return netw
				}
			}
		} else if strings.HasPrefix(cardNumber, iin) {
			return netw
		}
	}

	if strings.HasPrefix(cardNumber, "4") {
		return "Visa"
	}

	return ""
}

func handleRequest(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	log.Printf("Received request: %+v", request)
	var req cardNumberRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		return events.LambdaFunctionURLResponse{Body: err.Error(), StatusCode: 400}, err
	}
	rp := strings.NewReplacer(" ", "", "-", "")
	cardNumber := rp.Replace(req.CardNumber)
	isNumeric := regexp.MustCompile(`^[0-9]+$`).MatchString(cardNumber)
	if len(cardNumber) < 8 || len(cardNumber) > 19 || !isNumeric {
		log.Printf("Invalid card number: %v", req.CardNumber[:4])
		err := fmt.Errorf("invalid card number: %v", req.CardNumber[:4])
		return events.LambdaFunctionURLResponse{Body: err.Error(), StatusCode: 400}, err
	}

	resp := cardNumberResponse{
		cardNumberRequest: req,
		Valid:             isValidCardNumber(cardNumber),
		Network:           getNetWorkName(cardNumber),
	}

	b, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return events.LambdaFunctionURLResponse{Body: err.Error(), StatusCode: 400}, err
	}
	return events.LambdaFunctionURLResponse{Body: string(b), StatusCode: 200}, nil
}

func main() {
	lambda.Start(handleRequest)
}

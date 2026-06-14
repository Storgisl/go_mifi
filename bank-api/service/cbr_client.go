package service

import (
    "bytes"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/beevik/etree"
    "github.com/sirupsen/logrus"
)

type CBRClient struct{}

func NewCBRClient() *CBRClient {
    return &CBRClient{}
}

func (c *CBRClient) GetKeyRate() (float64, error) {
    fromDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
    toDate := time.Now().Format("2006-01-02")
    soapBody := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
        <soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
            <soap12:Body>
                <KeyRate xmlns="http://web.cbr.ru/">
                    <fromDate>%s</fromDate>
                    <ToDate>%s</ToDate>
                </KeyRate>
            </soap12:Body>
        </soap12:Envelope>`, fromDate, toDate)

    req, err := http.NewRequest("POST", "https://www.cbr.ru/DailyInfoWebServ/DailyInfo.asmx", bytes.NewBuffer([]byte(soapBody)))
    if err != nil {
        return 0, err
    }
    req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
    req.Header.Set("SOAPAction", "http://web.cbr.ru/KeyRate")

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return 0, err
    }

    doc := etree.NewDocument()
    if err := doc.ReadFromBytes(body); err != nil {
        logrus.WithError(err).Error("Failed to parse CBR XML")
        return 0, err
    }

    krElements := doc.FindElements("//diffgram/KeyRate/KR")
    if len(krElements) == 0 {
        return 0, fmt.Errorf("no key rate data")
    }
    rateElement := krElements[0].FindElement("./Rate")
    if rateElement == nil {
        return 0, fmt.Errorf("rate element missing")
    }
    var rate float64
    if _, err := fmt.Sscanf(rateElement.Text(), "%f", &rate); err != nil {
        return 0, err
    }
    rate += 5.0
    logrus.Infof("Current CBR key rate + margin: %.2f", rate)
    return rate, nil
}

package main

import (
	"encoding/json"
    "encoding/base64"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"
    "github.com/btcsuite/btcutil/bech32"
    "golang.org/x/text/message"
    "golang.org/x/text/language"
)

const (
    icnsRest = "https://lcd.osmosis.zone/cosmwasm/wasm/v1/contract/osmo1xk0s8xgktn9x5vwcgtjdxqzadg88fgn33p8u9cnpdxwemvxscvast52cdd/smart/"

    fundRestTx = "https://rest.unification.io/cosmos/tx/v1beta1/txs/"
    fundExplorerTx = "https://explorer.unification.io/transactions/"

    fundExplorerAccount = "https://explorer.unification.io/accounts/"
    osmoExplorerAccount = "https://www.mintscan.io/osmosis/address/"
    gravExplorerAccount = "https://www.mintscan.io/gravity-bridge/address/"
)

// Places a string in HTML bold brackets
func mkBold(msg string) string{
    return fmt.Sprintf("<b>%s</b>",msg)
}

// Returns and HTML formatted hyperlink for an account when given a wallet or validator address
func mkAccountLink(addr string) string{
    switch addr[:6]{
    case "undval":
        return fmt.Sprintf("<a href=\"%s%s\">%s</a>",fundExplorerValidators,addr,getAccountName(addr))
    }
    switch addr[:3]{
    case "osm":
        return fmt.Sprintf("<a href=\"%s%s\">%s</a>",osmoExplorerAccount,addr,getAccountName(addr))
    case "gra":
        return fmt.Sprintf("<a href=\"%s%s\">%s</a>",gravExplorerAccount,addr,getAccountName(addr))
    default:
        return fmt.Sprintf("<a href=\"%s%s\">%s</a>",fundExplorerAccount,addr,getAccountName(addr))
    }
}

// Returns a HTML formatted hyprlink for a transaction when given a TX Hash with an amount
func mkTranscationLink(hash string, amount string) string {
    return fmt.Sprintf("<a href=\"%s%s\">%s</a>",fundExplorerTx,hash,denomToAmount(amount))
}

// When given a transaction hash
// Searches rest endpoints for a memo on the transaction, if not available returns an empty string
func getMemo(hash string) string {
    var tx TxResponse
    resp, err := http.Get(fundRestTx + hash)
    if err != nil {
        log.Println("Could TX rest response: ", err)
        return ""
    }
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Println("Failed to read TX rest response: ", err)
        return ""
    }
    if err := json.Unmarshal(body, &tx); err != nil {
        log.Println("Failed to unmarshal TX rest response: ", err)
        return ""
    }
    return tx.Tx.Body.Memo
}


// When given a wallet or validator address, returns the name associated with the wallet, if it has one
// Otherwise returns a truncated version of the wallet address
func getAccountName(msg string) string {

    // Known account names
    named := map[string][]string{
        "BitForex 🏦": {"und18mcmhkq6fmhu9hpy3sx5cugqwv6z0wrz7nn5d7", ""},
        "Poloniex 🏦" : {"und186slma7kkxlghwc3hzjr9gkqwhefhln5pw5k26",""},
        "ProBit 🏦" : {"und1jkhkllr3ws3uxclawn4kpuuglffg327wvfg8r9",""},
        "DigiFinex 🏦" : {"und1xnrruk9qlgnmh8qxcz9ypfezj45qk96v2rgnzk",""},
        "All Unjailed Delegations" : {"und1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3j7wxl3",""},
        "Burn Address 🔥" : {"und1qqqqqqqqqqqqqqqqqqqqqqqqqqqqph4djz5txt",""},
        "Unbonding/Jailed Delegations" : {"und1tygms3xhhs3yv487phx3dw4a95jn7t7lx7jhf9",""},
        "Locked eFUND" : {"und1nwt6chnk0efe8ngwa5y63egmdumht6arlvluh3",""},
        "wFUND" : {"und12k2pyuylm9t7ugdvz67h9pg4gmmvhn5vcrzmhj",""},
        "Foundation Wallet #1\n( Cold Wallet ) 🏛️" : {"und1fxnqz9evaug5m4xuh68s62qg9f5xe2vzsj44l8",""},
        "Foundation Wallet #2 🏛️" : {"und1pyqttnfyqujh4hvjhcx45mz8svptp6f40n4u3p",""},
        "Foundation Wallet #3 🏛️" : {"und1hdn830wndtquqxzaz3rds7r7hqgpsg5q9ggxpk",""},
        "Foundation Wallet #4 🏛️" : {"und1cwhkh2ag8w2lf3ngd509wzy43ljxkkn3qe3q4z",""},
    }
    // Convert undval to und1 addresses and append to map
    for _, val := range vals.Validators {
        _, data, _ := bech32.Decode(val.OperatorAddress)
        addr, _ := bech32.Encode("und",data)
        named[val.Description.Moniker] = []string{addr, val.OperatorAddress}
    }
    // Check if name matches wallet or val addr
    for key, val := range named {
        if val[0] == msg || val[1] == msg {
            return key
        }
    }

    // Check ICNS for name
    var icns ICNS
    query := fmt.Sprintf(`{ "primary_name": { "address": "%s" }}`, msg)
    b64 := base64.StdEncoding.EncodeToString([]byte(query))
    resp, err := http.Get("https://lcd.osmosis.zone/cosmwasm/wasm/v1/contract/osmo1xk0s8xgktn9x5vwcgtjdxqzadg88fgn33p8u9cnpdxwemvxscvast52cdd/smart/" + b64); 
    if err != nil {
        log.Println("Failed to get ICNS Response")
    } 
    defer resp.Body.Close()
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        fmt.Println("Failed to read ICNS response")
    }
    if err := json.Unmarshal(body, &icns); err != nil {
        fmt.Println("Failed to unmarshal ICNS response")
    }
    if icns.Data.Name != "" {
        return icns.Data.Name + " (ICNS)"
    }

    // Return truncated addr if the addr isnt in the named map
    return fmt.Sprintf("%s...%s",msg[:7],msg[len(msg)-7:])
}



// Converts the denom to the formatted amount
// E.G. 1000000000nund becomes 1.00 FUND
func denomToAmount(msg string) string {
    var amount string
    var denom string

    switch msg[len(msg)-4:] {
    case "nund":
        denom = "nund"
        amount = msg[:len(msg)-4]
    default:
        // Other IBC denoms such as ibc/xxxx
        // IBC denom hash is always 64 chars + 4 chars for the ibc/
        denom = msg[len(msg)-68:]
        amount = msg[:len(msg)-68]
    }

    numericalAmount, _ := strconv.ParseFloat(amount, 64)
    // This will format the numbers in human readable form E.G. 1000 FUND should become 1,000 FUND
    formatter := message.NewPrinter(language.English)

    switch denom {
    case "nund":
        // Fund
        numericalAmount = math.Round((numericalAmount/1000000000)*100)/100
        return formatter.Sprintf("%.2f FUND ($%.2f USD)", numericalAmount, (cg.MarketData.CurrentPrice.USD * numericalAmount))

    case "ibc/ED07A3391A112B175915CD8FAF43A2DA8E4790EDE12566649D0C2F97716B8518":
        // Osmo
        numericalAmount = math.Round((numericalAmount/1000000)*100)/100
        return formatter.Sprintf("%.2f FUND ($%.2f USD)", numericalAmount)

    default:
        return "Unknown IBC"
    }
}


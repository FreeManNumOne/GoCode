package main
// prometheus exporter客户端
import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// 1. 创建自定义的 Prometheus 指标
var (
	requestCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "myapp_http_requests_total",
			Help: "Total number of HTTP requests",
		},
	)

	requestLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "myapp_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)
)

func init() {
	// 2. 注册指标
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestLatency)
}

// 模拟的处理请求的函数
func handleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// 3. 增加请求计数
	requestCount.Inc()

	// 模拟随机延迟
	delay := rand.Intn(2000)
	time.Sleep(time.Duration(delay) * time.Millisecond)

	w.Write([]byte("Hello, Prometheus!"))

	// 4. 记录请求处理时间
	duration := time.Since(start).Seconds()
	requestLatency.Observe(duration)
}

func main() {
	// 5. 暴露 /metrics 端点，Prometheus 将从这里抓取数据
	http.Handle("/metrics", promhttp.Handler())
	// 创建一个简单的 HTTP 服务端，响应其他路径请求
	http.HandleFunc("/", handleRequest)
	log.Println("Starting server on :8080")
	// 6. 启动 HTTP 服务器
	log.Fatal(http.ListenAndServe(":8080", nil))
}


Go 沒有真正的全局鎖的概念，因為它鼓勵輕量級的併發和 goroutine 的使用。使用全局互斥鎖會導致嚴重的性能瓶頸，因為所有 goroutine 都需要爭奪同一個鎖。  然而，你可以模擬全局鎖的行為，但通常不建議這樣做。

以下是一些方法來模擬全局鎖，並討論它們的優缺點：

**方法一：使用 `sync.Mutex`**

這是最直接的方法，但也是最容易造成性能問題的方法。  所有需要訪問受保護資源的 goroutine 都必須獲得這個 `Mutex`。

```go
package main

import (
    "fmt"
    "sync"
)

var (
    globalMutex sync.Mutex
    globalData  int
)

func incrementGlobalData() {
    globalMutex.Lock()
    defer globalMutex.Unlock()
    globalData++
}

func main() {
    for i := 0; i < 1000; i++ {
        go incrementGlobalData()
    }

    // 等待所有 goroutine 完成 (不精確，僅供演示)
    //  更好的方式是使用WaitGroup
    time.Sleep(100 * time.Millisecond)

    fmt.Println("Final globalData:", globalData)
}

```

**缺點：**  這會創建一個單一的瓶頸點。所有 goroutine 都必須串行化，抵消了 Go 並發的優點。  在大規模併發下，性能會非常差。

**方法二：使用 `sync.RWMutex` (讀寫鎖)**

如果你的全局資源大部分時間都是讀取的，只有小部分時間需要寫入，那麼 `sync.RWMutex` 可以提高性能。  多個 goroutine 可以同時讀取，但寫入時需要獨佔鎖。

```go
package main

import (
    "fmt"
    "sync"
)

var (
    globalRWMutex sync.RWMutex
    globalData    int
)

func readGlobalData() int {
    globalRWMutex.RLock()
    defer globalRWMutex.RUnlock()
    return globalData
}

func incrementGlobalData() {
    globalRWMutex.Lock()
    defer globalRWMutex.Unlock()
    globalData++
}

func main() {
    // ... (Similar to the previous example)
}
```

**缺點：**  仍然存在鎖競爭的可能性，只是比 `sync.Mutex` 稍微好一些。

**方法三：使用通道 (Channel)**

通道提供了一種更 Go 風格的同步機制。  你可以使用一個單向通道來控制對全局資源的訪問。

```go
package main

import (
    "fmt"
    "sync"
)

var (
    globalData chan int
    wg sync.WaitGroup
)

func init() {
    globalData = make(chan int, 1) // Buffered channel to allow for some concurrency
    globalData <- 0 // Initialize with 0
}

func incrementGlobalData() {
    defer wg.Done()
    for {
        value := <-globalData
        value++
        globalData <- value
        break;
    }
}

func main() {
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go incrementGlobalData()
    }
    wg.Wait()

    // read from the channel
    finalValue := <-globalData
    fmt.Println("Final globalData:", finalValue)
}
```

**優點：**  更符合 Go 的風格，避免了顯式的鎖定。

**缺點：**  需要謹慎設計通道的容量，避免阻塞。


**建議：**

在絕大多數情況下，避免使用全局鎖。  重新設計你的程式碼，將全局狀態分解成更小的、可併發訪問的單元。  考慮使用更細粒度的鎖，或者使用無鎖數據結構，例如原子操作，以提高性能和可擴展性。  只有在絕對必要的情況下才考慮使用全局鎖，並且要仔細評估其性能影響。  通道通常是更好的選擇，因為它們更符合 Go 的併發模型。


記住，選擇哪種方法取決於你的具體需求。  仔細考慮你的程式碼的併發模型和性能需求，選擇最適合你的方法。  使用 `sync.WaitGroup` 確保所有 goroutine 完成是良好的實踐。  仔細測試你的程式碼，以確保它在高併發情況下也能正確工作。
WARNING: All log messages before absl::InitializeLog() is called are written to STDERR
E0000 00:00:1739257475.697419 114526637 init.cc:232] grpc_wait_for_shutdown_with_timeout() timed out.

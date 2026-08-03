# Channel trong Go

## 1. Channel là gì?

Channel là "đường ống" (pipe) để các goroutine **giao tiếp và đồng bộ dữ liệu** với nhau một cách an toàn (thread-safe), theo triết lý của Go:

> "Don't communicate by sharing memory; share memory by communicating."

Thay vì dùng biến chung + mutex, các goroutine gửi dữ liệu cho nhau qua channel.

## 2. Khai báo và khởi tạo

```go
ch := make(chan Course)        // unbuffered channel (không có buffer)
ch := make(chan Message, 10)   // buffered channel (buffer chứa tối đa 10 phần tử)
```

- Channel có **kiểu dữ liệu** cụ thể: `chan Course` chỉ gửi/nhận được `Course`.
- Phải tạo bằng `make` trước khi dùng. Channel `nil` sẽ block vĩnh viễn khi gửi/nhận.

## 3. Gửi và nhận dữ liệu

```go
ch <- course      // gửi (send) dữ liệu vào channel
c := <-ch         // nhận (receive) dữ liệu từ channel
```

Ví dụ cơ bản trong [main.go](main.go) (phần comment):

```go
ch := make(chan Course)

go func() {
    course := Course{Title: "Tips 30", Price: 500}
    ch <- course // goroutine gửi dữ liệu
}()

c := <-ch // main goroutine nhận dữ liệu (block cho đến khi có dữ liệu)
```

### Tính chất blocking (quan trọng nhất)

Với **unbuffered channel**:
- `ch <- x` **block** cho đến khi có goroutine khác nhận.
- `<-ch` **block** cho đến khi có goroutine khác gửi.

Đây cũng chính là cơ chế **đồng bộ**: ở ví dụ trên, `main` đứng chờ ở `<-ch` nên chương trình không kết thúc trước khi goroutine gửi xong.

Với **buffered channel**: gửi chỉ block khi buffer **đầy**, nhận chỉ block khi buffer **rỗng**.

## 4. Directional channel (channel một chiều)

Dùng trong tham số hàm để giới hạn quyền, giúp compiler bắt lỗi sớm:

```go
func publisher(channel chan<- Message, orders []Message) // chan<- : CHỈ ĐƯỢC GỬI
func subcribe(channel <-chan Message, userName string)   // <-chan : CHỈ ĐƯỢC NHẬN
```

- `chan<- Message`: send-only — trong `publisher` nếu viết `<-channel` sẽ bị lỗi compile.
- `<-chan Message`: receive-only — trong `subcribe` không thể gửi hay `close`.
- Channel hai chiều (`chan Message`) tự động convert sang một chiều khi truyền vào hàm.

## 5. `close` và `range` trên channel

```go
func publisher(channel chan<- Message, orders []Message) {
    for _, order := range orders {
        channel <- order
    }
    close(channel) // báo hiệu: không còn dữ liệu nữa
}

func subcribe(channel <-chan Message, userName string) {
    for msg := range channel { // tự động dừng khi channel bị close
        fmt.Printf("usr %s::Orders:%s:: Title:%s", userName, msg.OrderId, msg.Title)
    }
}
```

Quy tắc về `close`:
- **Chỉ bên gửi (sender) mới nên close** channel, không bao giờ là bên nhận.
- Gửi vào channel đã close → **panic**.
- Nhận từ channel đã close → trả về ngay **zero value**, không block.
- Kiểm tra channel còn mở hay không: `msg, ok := <-ch` (`ok == false` nghĩa là đã close và hết dữ liệu).
- `for range` trên channel sẽ tự thoát khi channel close — nếu quên `close`, vòng lặp block mãi → **goroutine leak / deadlock**.

## 6. Mô hình Pub/Sub trong main.go

```
publisher goroutine ──(Message)──> orderChannel ──(Message)──> subscriber goroutine
```

- `publisher`: đẩy từng order vào channel rồi `close`.
- `subcribe`: dùng `for range` đọc đến khi channel close.

### ⚠️ Bug trong code hiện tại: main thoát trước khi goroutine chạy xong

```go
func main() {
    orderChannel := make(chan Message)
    orders := []Message{{Title: "sach", OrderId: "123"}}

    go publisher(orderChannel, orders)
    go subcribe(orderChannel, "Lam Nguyen 10")
    // main kết thúc NGAY tại đây → toàn bộ goroutine bị kill
    // → gần như không in ra gì cả!
}
```

Khi hàm `main` return, chương trình kết thúc **không chờ** các goroutine khác. Cách sửa phổ biến nhất là dùng `sync.WaitGroup`:

```go
func main() {
    orderChannel := make(chan Message)
    orders := []Message{{Title: "sach", OrderId: "123"}}

    var wg sync.WaitGroup
    wg.Add(1)

    go publisher(orderChannel, orders)
    go func() {
        defer wg.Done()
        subcribe(orderChannel, "Lam Nguyen 10")
    }()

    wg.Wait() // chờ subscriber đọc hết dữ liệu
}
```

(Hoặc đơn giản hơn với 1 subscriber: chạy `subcribe` trực tiếp trong main thay vì `go subcribe(...)` — main sẽ tự block ở `for range` đến khi channel close.)

## 7. Các lỗi thường gặp (tóm tắt)

| Lỗi | Nguyên nhân | Hậu quả |
|---|---|---|
| Deadlock | Gửi/nhận unbuffered channel mà không có goroutine đối ứng | `fatal error: all goroutines are asleep - deadlock!` |
| Panic | Gửi vào channel đã close, hoặc close 2 lần | `panic: send on closed channel` |
| Goroutine leak | `for range` channel nhưng sender quên `close` | Goroutine block vĩnh viễn, tốn bộ nhớ |
| Mất output | `main` return trước khi goroutine chạy xong | Không in gì (bug ở mục 6) |

## 8. Kiến thức mở rộng nên học tiếp

- **`select`**: chờ trên nhiều channel cùng lúc, kết hợp `time.After` để timeout.
- **Buffered channel**: `make(chan T, n)` — giảm blocking, làm hàng đợi.
- **Worker pool**: nhiều goroutine cùng đọc 1 channel để chia việc.
- **`context.Context`**: hủy goroutine một cách chủ động.
- Lưu ý: pub/sub thật sự (1 message đến **nhiều** subscriber) cần fan-out — với 1 channel, nhiều subscriber sẽ **chia nhau** message chứ không phải ai cũng nhận được (đây là work distribution, không phải broadcast).

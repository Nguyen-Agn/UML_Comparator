import json
from http.server import HTTPServer, BaseHTTPRequestHandler

class MockOpenAIHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        if self.path == '/v1/chat/completions':
            # Đọc body để giả lập (có thể in ra nếu muốn xem)
            content_length = int(self.headers.get('Content-Length', 0))
            if content_length > 0:
                self.rfile.read(content_length)
                print("Da nhan request thanh cong, dang tra ve mock data...")

            self.send_response(200)
            self.send_header('Content-Type', 'application/json')
            self.send_header('Access-Control-Allow-Origin', '*')
            self.end_headers()
            
            # Mock data trả về (Một biểu đồ UML Mermaid hợp lệ)
            mock_mermaid = """```mermaid
classDiagram
    Store "1" *-- "*" Order : __1__
    Order "1" *-- "*" Product : __1__

    class Store {
      + "name|title" : "String|char[]" Readonly __1__
      + "manageInventory()" "void" __1__
    }

    class Order {
      - "orderId|Id|id" : "int|Integer" __1__
      - "totalAmount|totalAmount" : "double|Double" __1__
      + "processPayment(method: String)" "boolean|Boolean" __1__
    }

    class Product {
      - "productId|Id|IP|ip" : "int|Integer" __1__
      + "Price|price|P|cost" : "double|Double" __1__
      + "updatePrice(newPrice: double)" "void|boolean" __1__
    }
```"""
            
            response = {
                "choices": [
                    {
                        "message": {
                            "role": "assistant",
                            "content": mock_mermaid
                        }
                    }
                ]
            }
            self.wfile.write(json.dumps(response).encode('utf-8'))
        else:
            self.send_response(404)
            self.end_headers()

if __name__ == '__main__':
    port = 8081
    server_address = ('', port)
    httpd = HTTPServer(server_address, MockOpenAIHandler)
    print(f"Mock OpenAI Server dang chay tai: http://localhost:{port}")
    print(f"Trong app, hay cau hinh API Endpoint la: http://localhost:{port}/v1")
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\nĐã dừng server.")

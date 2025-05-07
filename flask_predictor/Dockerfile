# Base image
FROM python:3.10

# Tạo thư mục làm việc
WORKDIR /app

# Chỉ copy requirements trước để cache cài thư viện
COPY requirements.txt .

# Cài đặt thư viện trước
RUN pip install --no-cache-dir -r requirements.txt

# Sau đó mới copy mã nguồn (code thay đổi thì không ảnh hưởng đến bước pip install)
COPY . .

# Mở port Flask
EXPOSE 5000

# Chạy app Flask
CMD ["python", "app.py"]

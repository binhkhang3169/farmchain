FROM node:22.14-alpine

# Tạo thư mục làm việc
WORKDIR /app

# Sao chép package.json và cài dependencies trước (tận dụng cache Docker)
COPY package*.json ./

# Cài dependencies
RUN npm install

# Sao chép toàn bộ mã nguồn vào container
COPY . .

# Mở cổng mặc định của dev server (có thể thay 3000 nếu bạn dùng cổng khác)
EXPOSE 3000

# Chạy lệnh dev
CMD ["npm", "run", "dev"]

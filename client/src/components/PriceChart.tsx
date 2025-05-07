import { Line } from 'react-chartjs-2';
import React, { useState, useEffect } from 'react';
import {
  Chart as ChartJS,
  LineElement,
  PointElement,
  LinearScale,
  CategoryScale,
  Title,
  Tooltip,
  Legend,
  ChartData,
  ChartOptions,
} from 'chart.js';

ChartJS.register(
  LineElement, 
  PointElement, 
  LinearScale, 
  CategoryScale, 
  Title, 
  Tooltip, 
  Legend
);

// Định nghĩa kiểu dữ liệu
interface HistoricalPrice {
  day: string;
  price: number;
}

interface PredictedPrice {
  day: string;
  price: number;
}

interface PriceData {
  history: HistoricalPrice[];
  predictions: PredictedPrice[];
}

// Kiểu dữ liệu cho chart
interface PriceChartData {
  labels: string[];
  datasets: {
    label: string;
    data: (number | null)[];
    fill: boolean;
    borderColor: string;
    tension: number;
    pointRadius?: number;
    pointStyle?: string;
    pointBackgroundColor?: string;
  }[];
}

const PriceChart: React.FC = () => {
  // Định nghĩa kiểu dữ liệu cho state
  const [chartData, setChartData] = useState<PriceChartData>({
    labels: [],
    datasets: [
      {
        label: 'Giá lịch sử',
        data: [],
        fill: false,
        borderColor: 'rgb(59,130,246)',
        tension: 0.1,
        pointRadius: 3,
      },
      {
        label: 'Dự đoán giá',
        data: [], // Sẽ được điền khi có dữ liệu
        fill: false,
        borderColor: 'rgb(220,53,69)',
        tension: 0.1,
        pointStyle: 'star',
        pointRadius: 5,
        pointBackgroundColor: 'rgb(220,53,69)',
      },
    ],
  });
  
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchPriceData();
  }, []);

  // Định dạng ngày từ 'YYYY-MM-DD' thành 'DD/MM'
  const formatDate = (dateString: string): string => {
    const date = new Date(dateString);
    return `${date.getDate()}/${date.getMonth() + 1}`;
  };

  const fetchPriceData = async (): Promise<void> => {
    try {
      setIsLoading(true);
      // Gọi API từ Go server
      const response = await fetch('http://localhost:8080/api/price-history');
      
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
      
      const data: PriceData = await response.json();
      
      // Tạo mảng cho nhãn và dữ liệu với kiểu xác định
      const labels: string[] = [];
      const historicalPrices: (number | null)[] = [];
      const predictedPrices: (number | null)[] = [];
      
      // Lấy dữ liệu lịch sử
      data.history.forEach((item: HistoricalPrice) => {
        labels.push(formatDate(item.day));
        historicalPrices.push(item.price);
        predictedPrices.push(null); // Thêm null cho dự đoán ở vị trí tương ứng
      });
      
      // Số điểm dữ liệu lịch sử
      const historyCount: number = labels.length;
      
      // Thêm dữ liệu dự đoán
      data.predictions.forEach((item: PredictedPrice) => {
        labels.push(formatDate(item.day));
        historicalPrices.push(null); // Thêm null cho lịch sử ở vị trí tương ứng
        predictedPrices.push(item.price);
      });
      
      // Chỉnh sửa giá trị null cho overlap
      if (historyCount > 0 && data.predictions.length > 0) {
        // Kiểm tra kiểu dữ liệu trước khi gán
        const lastHistoricalPrice = historicalPrices[historyCount - 1];
        if (lastHistoricalPrice !== null) {
          predictedPrices[historyCount - 1] = lastHistoricalPrice; // Gỡ null ở điểm kết nối
        }
      }
      
      // Cập nhật dữ liệu biểu đồ
      setChartData({
        labels,
        datasets: [
          {
            ...chartData.datasets[0],
            data: historicalPrices,
          },
          {
            ...chartData.datasets[1],
            data: predictedPrices,
          },
        ],
      });
      
    } catch (err) {
      console.error("Error fetching price data:", err);
      setError(err instanceof Error ? err.message : "Unknown error");
    } finally {
      setIsLoading(false);
    }
  };

  // Định nghĩa kiểu cho options
  const options: ChartOptions<'line'> = {
    responsive: true,
    plugins: {
      title: {
        display: true,
        text: 'Biểu đồ giá sản phẩm',
        font: {
          size: 16,
        },
      },
      legend: {
        position: 'bottom',
      },
      tooltip: {
        callbacks: {
          title: function(context) {
            return `day: ${context[0].label}`;
          },
          label: function(context) {
            if (context.parsed.y !== null) {
              return `${context.dataset.label}: ${context.parsed.y.toLocaleString('vi-VN')} VND`;
            }
            return '';
          }
        }
      }
    },
    interaction: {
      intersect: false,
      mode: 'index',
    },
    scales: {
      x: {
        title: {
          display: true,
          text: 'Ngày'
        }
      },
      y: {
        title: {
          display: true,
          text: 'Giá (VND)'
        },
        beginAtZero: false,
        ticks: {
          // Định dạng số với đơn vị tiền tệ Việt Nam
          callback: function(value) {
            return value.toLocaleString('vi-VN');
          }
        }
      }
    }
  };

  if (isLoading) {
    return (
      <div className="bg-white border rounded p-4 w-full flex justify-center items-center h-64">
        <div className="text-center">
          <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-500 mx-auto"></div>
          <p className="mt-2 text-gray-600">Đang tải dữ liệu...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white border rounded p-4 w-full flex justify-center items-center h-64">
        <div className="text-center text-red-500">
          <p>Có lỗi xảy ra khi tải dữ liệu: {error}</p>
          <button 
            className="mt-2 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
            onClick={fetchPriceData}
          >
            Thử lại
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white border rounded p-4 w-full">
      <Line data={chartData as ChartData<'line'>} options={options} />
      <div className="mt-4 text-sm text-gray-600">
        <p>* Dự đoán giá được tính dựa trên dữ liệu lịch sử và mô hình AI</p>
      </div>
    </div>
  );
};

export default PriceChart;
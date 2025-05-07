import { useNavigate } from 'react-router-dom';

function App() {
  const navigate = useNavigate();

  const handleRoleSelect = (role: string) => {
    if (role === 'buyer') {
      navigate('/map/buyer');
    } else if (role === 'seller') {
      navigate('/map/seller');
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <div className="bg-white p-8 rounded shadow text-center space-y-4">
        <h1 className="text-xl font-semibold">NongSanNET - Chợ Nông Sản Trực Tuyến</h1>
        <p className="text-gray-600">Chào mừng đến với nền tảng kết nối người mua và người bán nông sản</p>
        <h2 className="text-lg font-medium mt-4">Chọn vai trò của bạn</h2>
        <div className="flex space-x-4">
          <button
            onClick={() => handleRoleSelect('buyer')}
            className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
          >
            Người Mua
          </button>
          <button
            onClick={() => handleRoleSelect('seller')}
            className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600"
          >
            Người Bán
          </button>
        </div>
      </div>
    </div>
  );
}

export default App;
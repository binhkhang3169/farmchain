import { useNavigate } from 'react-router-dom';
import ChatBox from '../components/ChatBox';
import PriceChart from '../components/PriceChart';

interface NegotiationPageProps {
  role: string;
}

const NegotiationPage = ({ role }: NegotiationPageProps) => {
  const navigate = useNavigate();

  const handleBackToMap = () => {
    if (role === 'buyer') {
      navigate('/map/buyer');
    } else {
      navigate('/map/seller');
    }
  };

  return (
    <div className="min-h-screen bg-gray-100 p-6">
      <div className="mb-4">
        <button 
          onClick={handleBackToMap}
          className="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600"
        >
          ← Quay lại bản đồ
        </button>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <PriceChart />
        <ChatBox role={role} />
      </div>
    </div>
  );
};

export default NegotiationPage;
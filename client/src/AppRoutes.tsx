import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import App from './App';
import BuyerMapPage from './pages/BuyerMapPage';
import SellerMapPage from './pages/SellerMapPage';
import NegotiationPage from './pages/NegotiationPage';

const AppRoutes = () => {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/map/buyer" element={<BuyerMapPage />} />
        <Route path="/map/seller" element={<SellerMapPage />} />
        <Route path="/negotiate/buyer" element={<NegotiationPage role="buyer" />} />
        <Route path="/negotiate/seller" element={<NegotiationPage role="seller" />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Router>
  );
};

export default AppRoutes;
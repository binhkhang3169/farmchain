import { useEffect, useState } from 'react';
import { MapContainer, TileLayer, Polygon, Popup } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import { useNavigate } from 'react-router-dom';

// Cấu hình lại biểu tượng mặc định
delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
    iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
    iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
    shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
});

interface AgriInfo {
    id: string;
    cropType: string;
    plantingDate: string;
    harvestDate: string;
    geoJson: string;
}

const BuyerMapPage = () => {
    const [areas, setAreas] = useState<AgriInfo[]>([]);
    const [selectedArea, setSelectedArea] = useState<AgriInfo | null>(null);
    const [mapInstance, setMapInstance] = useState<L.Map | null>(null);
    const navigate = useNavigate();

    useEffect(() => {
        const fetchAreas = async () => {
            try {
                const response = await fetch('http://localhost:8080/areas');
                const data = await response.json();
                setAreas(data);
            } catch (error) {
                console.error('Lỗi khi lấy dữ liệu:', error);
            }
        };
        fetchAreas();
    }, []);

    useEffect(() => {
        if (navigator.geolocation && mapInstance) {
            navigator.geolocation.getCurrentPosition((position) => {
                const { latitude, longitude } = position.coords;
                mapInstance.setView([latitude, longitude], 12);
            });
        }
    }, [mapInstance]);

    const handleAreaClick = (area: AgriInfo) => {
        setSelectedArea(area);
        try {
            const geoJson = JSON.parse(area.geoJson);
            if (geoJson.type === 'Polygon') {
                const bounds = L.latLngBounds(
                    geoJson.coordinates[0].map((coords: [number, number]) => [coords[1], coords[0]])
                );
                mapInstance?.fitBounds(bounds);
            }
        } catch (err) {
            console.error('Lỗi xử lý GeoJSON:', err);
        }
    };

    const handleStartNegotiation = () => {
        if (selectedArea) {
            navigate(`/negotiate/buyer?id=${selectedArea.id}`);
        }
    };

    return (
        <div className="relative h-screen w-screen">
            <MapContainer
                center={[10.762622, 106.660172]}
                zoom={6}
                style={{ height: '100%', width: '100%' }}
                whenReady={(map) => setMapInstance(map.target)}
            >
                <TileLayer
                    attribution="&copy; OpenStreetMap contributors"
                    url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                />

                {areas.map((area) => {
                    try {
                        const geoJson = JSON.parse(area.geoJson);
                        if (geoJson.type === 'Polygon') {
                            const positions = geoJson.coordinates[0].map(
                                (coords: [number, number]) => [coords[1], coords[0]]
                            );
                            return (
                                <Polygon
                                    key={area.id}
                                    positions={positions}
                                    pathOptions={{ color: 'green', weight: 2, opacity: 0.8 }}
                                    eventHandlers={{
                                        click: () => handleAreaClick(area),
                                    }}
                                >
                                    <Popup>
                                        <strong>🌾 Loại:</strong> {area.cropType}<br />
                                        <strong>📅 Trồng:</strong> {area.plantingDate}<br />
                                        <strong>🍂 Thu hoạch:</strong> {area.harvestDate}
                                    </Popup>
                                </Polygon>
                            );
                        }
                    } catch (err) {
                        console.error('Lỗi geoJson:', err);
                    }
                    return null;
                })}
            </MapContainer>

            {/* Panel thông tin vùng */}
            {selectedArea && (
                <div className="absolute top-4 right-4 bg-white p-4 rounded shadow-lg z-[1000] max-w-xs">
                    <h2 className="text-lg font-bold mb-2">Thông tin vùng</h2>
                    <div className="mb-2"><strong>🌾 Loại:</strong> {selectedArea.cropType}</div>
                    <div className="mb-2"><strong>📅 Trồng:</strong> {selectedArea.plantingDate}</div>
                    <div className="mb-2"><strong>🍂 Thu hoạch:</strong> {selectedArea.harvestDate}</div>
                    
                    <button
                        className="bg-blue-600 text-white w-full mt-2 py-2 rounded hover:bg-blue-700"
                        onClick={handleStartNegotiation}
                    >
                        Bắt đầu đàm phán
                    </button>
                </div>
            )}

            {!selectedArea && (
                <div className="absolute top-4 left-4 bg-white p-4 rounded shadow-lg z-[1000]">
                    <h2 className="text-lg font-bold mb-2">Hướng dẫn</h2>
                    <p>Chọn một vùng nông sản để xem thông tin và bắt đầu đàm phán.</p>
                </div>
            )}
        </div>
    );
};

export default BuyerMapPage;
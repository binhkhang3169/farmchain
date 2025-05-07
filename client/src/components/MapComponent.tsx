import { MapContainer, TileLayer, Polygon, FeatureGroup, Popup } from 'react-leaflet';
import { EditControl } from 'react-leaflet-draw';
import 'leaflet/dist/leaflet.css';
import 'leaflet-draw/dist/leaflet.draw.css';
import L from 'leaflet';
import { useRef, useState, useEffect } from 'react';
import * as turf from '@turf/turf';

delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
    iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
    iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
    shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
});

interface AgriInfo {
    cropType: string;
    plantingDate: string;
    harvestDate: string;
    id: string; // Mã vùng
}

const MapComponent: React.FC = () => {
    const featureGroupRef = useRef<L.FeatureGroup>(null);
    const [agriInfo, setAgriInfo] = useState<AgriInfo>({
        cropType: '',
        plantingDate: '',
        harvestDate: '',
        id: '',
    });
    const [isAreaSelected, setIsAreaSelected] = useState<boolean>(false);
    const [existingAreas, setExistingAreas] = useState<any[]>([]); // Lưu các vùng đã có
    const [mapInstance, setMapInstance] = useState<L.Map | null>(null);

    useEffect(() => {
        // Mô phỏng việc lấy dữ liệu vùng nông sản đã lưu
        // Ở thực tế, bạn sẽ gọi API hoặc lấy từ database
        const fetchAreas = async () => {
            try {
                const response = await fetch('http://localhost:8080/areas');
                const areas = await response.json();
                setExistingAreas(areas);
            } catch (error) {
                console.error('Error fetching areas:', error);
            }
        };

        fetchAreas();

        // Tự động lấy vị trí người dùng và zoom vào đó
        if (navigator.geolocation && mapInstance) {
            navigator.geolocation.getCurrentPosition((position) => {
                const { latitude, longitude } = position.coords;
                mapInstance.setView([latitude, longitude], 12);
            });
        }
    }, [mapInstance]);

    const handleSave = async () => {
        const layers = featureGroupRef.current?.getLayers();
        let geoJson = null;

        if (layers && layers.length > 0) {
            const layer = layers[0]; // Giả sử chỉ vẽ 1 polygon
            if (layer instanceof L.Polygon) {
                const latlngs = (layer.getLatLngs() as L.LatLng[][])[0];
                const coordinates = latlngs.map((latlng: L.LatLng) => [latlng.lng, latlng.lat]);

                // Đảm bảo polygon khép kín (nối điểm cuối với điểm đầu nếu chưa khép)
                if (
                    coordinates.length > 0 &&
                    (coordinates[0][0] !== coordinates[coordinates.length - 1][0] ||
                        coordinates[0][1] !== coordinates[coordinates.length - 1][1])
                ) {
                    coordinates.push([...coordinates[0]]);
                }

                geoJson = {
                    type: "Polygon",
                    coordinates: [coordinates],
                };
            }
        }

        const payload = {
            geoJson: JSON.stringify(geoJson),
            cropType: agriInfo.cropType,
            plantingDate: agriInfo.plantingDate,
            harvestDate: agriInfo.harvestDate,
        };

        try {
            const response = await fetch('http://localhost:8080/areas', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            });

            if (response.ok) {
                alert('Vùng nông sản đã được lưu thành công!');
            } else {
                alert('Có lỗi xảy ra!');
            }
        } catch (error) {
            console.error('Error:', error);
            alert('Không thể kết nối đến server!');
        }
    };



    const handleDrawEnd = (e: any) => {
        const layer = e.layer;

        // Convert Leaflet polygon to GeoJSON (turf format)
        const drawnGeoJSON = layer.toGeoJSON();
        const drawnPolygon = turf.polygon(drawnGeoJSON.geometry.coordinates);

        // Kiểm tra với tất cả các vùng đã có
        const isOverlap = existingAreas.some((area) => {
            const existing = JSON.parse(area.geoJson);
            const existingPolygon = turf.polygon(existing.coordinates);
            return turf.booleanOverlap(drawnPolygon, existingPolygon) || turf.booleanIntersects(drawnPolygon, existingPolygon);
        });

        if (isOverlap) {
            // Nếu trùng, xóa layer và cảnh báo
            const fg = featureGroupRef.current;
            if (fg) fg.removeLayer(layer);
            alert('🚫 Không thể vẽ chồng lên vùng đã có!');
            return;
        }

        // Nếu hợp lệ, cho phép nhập thông tin
        featureGroupRef.current?.addLayer(layer);
        setIsAreaSelected(true);
    };

    const handleAreaClick = (area: any) => {
        setAgriInfo({
            cropType: area.cropType,
            plantingDate: area.plantingDate,
            harvestDate: area.harvestDate,
            id: area.id,
        });

        // Zoom vào vùng được chọn
        if (mapInstance) {
            const geoJson = JSON.parse(area.geoJson);
            const bounds = L.latLngBounds(geoJson.coordinates[0].map((coords: [number, number]) => [coords[1], coords[0]]));
            mapInstance.fitBounds(bounds);
        }
    };

    return (
        <div className="relative">
            <h2 className="text-lg font-bold">Chọn vùng nông sản</h2>
            {isAreaSelected && (
                <div className="absolute top-0 left-0 w-full bg-white p-4 z-10">
                    <label className="block">
                        🌾 Loại nông sản:
                        <input
                            type="text"
                            className="border px-2 py-1 rounded w-full"
                            value={agriInfo.cropType}
                            onChange={(e) => setAgriInfo({ ...agriInfo, cropType: e.target.value })}
                        />
                    </label>
                    <label className="block">
                        📅 Thời gian trồng:
                        <input
                            type="date"
                            className="border px-2 py-1 rounded w-full"
                            value={agriInfo.plantingDate}
                            onChange={(e) => setAgriInfo({ ...agriInfo, plantingDate: e.target.value })}
                        />
                    </label>
                    <label className="block">
                        🍂 Thời gian thu hoạch:
                        <input
                            type="date"
                            className="border px-2 py-1 rounded w-full"
                            value={agriInfo.harvestDate}
                            onChange={(e) => setAgriInfo({ ...agriInfo, harvestDate: e.target.value })}
                        />
                    </label>
                    <button className="mt-2 bg-green-600 text-white px-4 py-2 rounded" onClick={handleSave}>
                        💾 Lưu vùng nông sản
                    </button>
                </div>
            )}

            <MapContainer
                center={[10.762622, 106.660172]}
                zoom={6}
                style={{ height: '100vh', width: '100%' }}
                whenReady={setMapInstance}
            >
                <TileLayer
                    attribution="&copy; OpenStreetMap contributors"
                    url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                />
                <FeatureGroup ref={featureGroupRef}>
                    <EditControl
                        position="topright"
                        draw={{
                            rectangle: false,
                            circle: false,
                            marker: false,
                            polyline: false,
                            circlemarker: false,
                        }}
                        onCreated={handleDrawEnd}
                    />
                </FeatureGroup>

                {existingAreas.map((area) => {
                    const geoJson = JSON.parse(area.geoJson);
                    return (
                        <Polygon
                            key={area.ID}
                            positions={geoJson.coordinates[0].map((coords: [number, number]) => [coords[1], coords[0]])}
                            color="blue"
                            weight={2}
                            opacity={0.7}
                            eventHandlers={{
                                click: () => handleAreaClick(area),
                            }}
                        >
                            <Popup>
                                <strong>Loại nông sản:</strong> {area.cropType}
                                <br />
                                <strong>Thời gian trồng:</strong> {area.plantingDate}
                                <br />
                                <strong>Thời gian thu hoạch:</strong> {area.harvestDate}
                            </Popup>
                        </Polygon>
                    );
                })}
            </MapContainer>
        </div>
    );
};

export default MapComponent;

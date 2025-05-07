interface ProductProps {
    name: string;
    image: string;
    price: number;
  }
  
  const ProductCard = ({ name, image, price }: ProductProps) => {
    return (
      <div className="border rounded p-4 w-full bg-white">
        <img src={image} alt={name} className="w-full h-48 object-cover rounded mb-2" />
        <h2 className="text-lg font-semibold">{name}</h2>
        <p className="text-gray-600">Giá đề xuất: ${price.toLocaleString()}</p>
      </div>
    );
  };
  
  export default ProductCard;
  
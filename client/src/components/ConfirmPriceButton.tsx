const ConfirmPriceButton = ({ onConfirm }: { onConfirm: () => void }) => {
    return (
      <button
        onClick={onConfirm}
        className="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700 w-full"
      >
        Xác nhận giá
      </button>
    );
  };

export default ConfirmPriceButton;
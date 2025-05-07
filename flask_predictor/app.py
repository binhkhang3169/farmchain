from flask import Flask, request, jsonify
import numpy as np
import pandas as pd
from tensorflow.keras.models import load_model
from sklearn.preprocessing import MinMaxScaler
import os
from datetime import datetime, timedelta
import logging

app = Flask(__name__)

# Logger
logging.basicConfig(level=logging.INFO)
logger = app.logger

# Cấu hình
MODEL_PATH = 'model/gru_model.h5'
DATA_PATH = 'data/data.xlsx'
WINDOW_SIZE = 30

def compute_indicators(df):
    logger.info("🔍 Tính chỉ báo kỹ thuật...")
    df['SMA'] = df['Giá'].rolling(window=5).mean()
    df['EMA'] = df['Giá'].ewm(span=5).mean()
    df['ROC'] = df['Giá'].pct_change(periods=5)
    delta = df['Giá'].diff()
    gain = (delta.where(delta > 0, 0)).rolling(window=14).mean()
    loss = (-delta.where(delta < 0, 0)).rolling(window=14).mean()
    rs = gain / (loss + 1e-6)
    df['RSI'] = 100 - (100 / (1 + rs))
    df['MACD'] = df['Giá'].ewm(span=12).mean() - df['Giá'].ewm(span=26).mean()
    return df.dropna()

def load_and_prepare_data():
    if not os.path.exists(DATA_PATH):
        logger.error("❌ Không tìm thấy dữ liệu.")
        return None, None, None

    try:
        df = pd.read_excel(DATA_PATH)
        df['Ngày'] = pd.to_datetime(df['Ngày'], format="%Y-%m-%d", errors='coerce')
        df = df.sort_values('Ngày')
        df['Giá'] = pd.to_numeric(df['Giá'], errors='coerce')
        df = df.dropna()

        df = compute_indicators(df).reset_index(drop=True)
        features = ['Giá', 'SMA', 'EMA', 'ROC', 'RSI', 'MACD']

        if len(df) < WINDOW_SIZE:
            logger.error("❌ Không đủ dữ liệu sau tính chỉ báo.")
            return None, None, None

        scaler = MinMaxScaler()
        df_scaled = scaler.fit_transform(df[features])
        df_scaled = pd.DataFrame(df_scaled, columns=features)

        return df, df_scaled, scaler

    except Exception as e:
        logger.exception("❌ Lỗi khi tải dữ liệu:")
        return None, None, None

def predict_next_days(df_raw, df_scaled, scaler, days=3):
    if not os.path.exists(MODEL_PATH):
        logger.error("❌ Không tìm thấy model.")
        return []

    try:
        model = load_model(MODEL_PATH, compile=False)
        features = ['Giá', 'SMA', 'EMA', 'ROC', 'RSI', 'MACD']

        last_date = df_raw['Ngày'].iloc[-1]
        current_data = df_scaled[-WINDOW_SIZE:].copy().values
        predictions = []

        for _ in range(days):
            X_input = current_data.reshape(1, WINDOW_SIZE, len(features))
            pred_scaled = model.predict(X_input, verbose=0)[0][0]

            dummy = np.zeros((1, len(features)))
            dummy[0, 0] = pred_scaled
            price_unscaled = scaler.inverse_transform(dummy)[0][0]

            next_date = last_date + timedelta(days=1)
            predictions.append({
                "day": next_date.strftime('%Y-%m-%d'),
                "price": round(float(price_unscaled), 2)
            })
            last_date = next_date

            # Cập nhật dữ liệu đầu vào
            new_row = current_data[-1].copy()
            new_row[0] = pred_scaled
            current_data = np.vstack([current_data[1:], new_row])

        return predictions

    except Exception as e:
        logger.exception("❌ Lỗi khi dự đoán:")
        return []

def predict_fair_price():
    logger.info("📦 Bắt đầu dự đoán giá hợp lý...")

    if not os.path.exists(MODEL_PATH):
        logger.error("❌ Không tìm thấy model.")
        return None, "Model not found"
    if not os.path.exists(DATA_PATH):
        logger.error("❌ Không tìm thấy dữ liệu.")
        return None, "Data not found"

    try:
        df = pd.read_excel(DATA_PATH)
        df['Ngày'] = pd.to_datetime(df['Ngày'])
        df['Giá'] = pd.to_numeric(df['Giá'], errors='coerce')
        df = df.dropna().sort_values('Ngày')
        df = compute_indicators(df).reset_index(drop=True)

        features = ['Giá', 'SMA', 'EMA', 'ROC', 'RSI', 'MACD']

        if len(df) < WINDOW_SIZE:
            return None, "Not enough data"

        scaler = MinMaxScaler()
        scaled = scaler.fit_transform(df[features])
        df_scaled = pd.DataFrame(scaled, columns=features)

        X_input = df_scaled[-WINDOW_SIZE:].values.reshape(1, WINDOW_SIZE, len(features))
        model = load_model(MODEL_PATH, compile=False)

        pred_scaled = model.predict(X_input)[0][0]
        dummy_input = np.zeros((1, len(features)))
        dummy_input[0, 0] = pred_scaled
        fair_price = scaler.inverse_transform(dummy_input)[0][0]

        return round(float(fair_price), 2), None

    except Exception as e:
        logger.exception("❌ Lỗi khi dự đoán giá hợp lý:")
        return None, str(e)

@app.route('/predict', methods=['POST'])
def predict():
    try:
        data = request.get_json()
        seller_price = float(data.get('seller_price', 0))
        buyer_price = float(data.get('buyer_price', 0))

        fair_price, error = predict_fair_price()
        if error:
            return jsonify({'error': error}), 500

        fair_price = round((fair_price * 0.5 + seller_price * 0.25 + buyer_price * 0.25), 2)

        if seller_price > fair_price:
            suggestion = f"Người bán nên giảm giá xuống gần {fair_price}"
        elif seller_price < fair_price:
            suggestion = f"Người bán có thể tăng giá lên gần {fair_price}"
        else:
            suggestion = "Giá người bán đang hợp lý!"

        return jsonify({
            'seller_price': seller_price,
            'buyer_price': buyer_price,
            'fair_price': fair_price,
            'suggestion': suggestion
        })

    except Exception as e:
        logger.exception("💥 Lỗi khi xử lý dự đoán:")
        return jsonify({'error': str(e)}), 500

@app.route('/price-history', methods=['GET'])
def get_price_history():
    try:
        df_raw, df_scaled, scaler = load_and_prepare_data()
        if df_raw is None or df_scaled is None:
            return jsonify({'error': 'Failed to load or prepare data'}), 500

        last_30_raw = df_raw.iloc[-30:].copy()
        history = [{
            "day": row["Ngày"].strftime("%Y-%m-%d"),
            "price": round(float(row["Giá"]), 2)
        } for _, row in last_30_raw.iterrows()]

        predictions = predict_next_days(df_raw, df_scaled, scaler, days=3)

        return jsonify({
            "history": history,
            "predictions": predictions
        })

    except Exception as e:
        logger.exception("⚠️ Lỗi trong /price-history:")
        return jsonify({'error': 'Internal server error'}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)

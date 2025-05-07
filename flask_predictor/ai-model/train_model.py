# ai-model/train_model.py

import pandas as pd
import numpy as np
import os
from datetime import datetime
from sklearn.preprocessing import MinMaxScaler
from tensorflow.keras.models import Sequential, load_model
from tensorflow.keras.layers import GRU, Dense, Dropout
from tensorflow.keras.callbacks import EarlyStopping, ReduceLROnPlateau
from tensorflow.keras.optimizers import Adam

WINDOW_SIZE = 30
MODEL_PATH = 'model/gru_model.h5'
DATA_PATH = 'data/historical_prices.csv'


# =============================
# 📈 Hàm tính chỉ số kỹ thuật
# =============================
def compute_indicators(df):
    df['SMA'] = df['Giá'].rolling(window=5).mean()
    df['EMA'] = df['Giá'].ewm(span=5, adjust=False).mean()
    df['ROC'] = df['Giá'].pct_change(periods=5)
    
    delta = df['Giá'].diff()
    gain = np.where(delta > 0, delta, 0)
    loss = np.where(delta < 0, -delta, 0)
    avg_gain = pd.Series(gain).rolling(window=14).mean()
    avg_loss = pd.Series(loss).rolling(window=14).mean()
    rs = avg_gain / (avg_loss + 1e-6)
    df['RSI'] = 100 - (100 / (1 + rs))
    
    df['MACD'] = df['Giá'].ewm(span=12).mean() - df['Giá'].ewm(span=26).mean()
    
    return df


# =============================
# 📊 Tiền xử lý dữ liệu
# =============================
def load_and_preprocess():
    df = pd.read_csv(DATA_PATH)
    df['Ngày'] = pd.to_datetime(df['Ngày'], format="%Y-%m-%d")
    df = df.sort_values('Ngày')
    df['Giá'] = pd.to_numeric(df['Giá'], errors='coerce')
    df = df.dropna()

    df = compute_indicators(df)
    df = df.dropna().reset_index(drop=True)

    features = ['Giá', 'SMA', 'EMA', 'ROC', 'RSI', 'MACD']
    scaler = MinMaxScaler()
    df_scaled = scaler.fit_transform(df[features])
    df_scaled = pd.DataFrame(df_scaled, columns=features)

    X, y = [], []
    for i in range(WINDOW_SIZE, len(df_scaled)):
        X.append(df_scaled.iloc[i - WINDOW_SIZE:i].values)
        y.append(df_scaled.iloc[i]['Giá'])

    X = np.array(X).reshape(-1, WINDOW_SIZE, len(features))
    y = np.array(y)
    return X, y, scaler


# =============================
# 🧠 Huấn luyện mô hình GRU
# =============================
def train_gru_model(X, y, continue_training=True):
    if os.path.exists(MODEL_PATH) and continue_training:
        print("📦 Đang load mô hình đã huấn luyện trước...")
        model = load_model(MODEL_PATH)
    else:
        print("✨ Tạo mô hình mới...")
        model = Sequential([
            GRU(64, return_sequences=False, input_shape=(X.shape[1], X.shape[2])),
            Dropout(0.2),
            Dense(1)
        ])
        model.compile(optimizer=Adam(learning_rate=0.001), loss='mse')

    callbacks = [
        EarlyStopping(monitor='loss', patience=5, restore_best_weights=True),
        ReduceLROnPlateau(monitor='loss', factor=0.5, patience=3)
    ]

    model.fit(X, y, epochs=50, batch_size=16, callbacks=callbacks, verbose=1)
    model.save(MODEL_PATH)
    print(f"✅ Mô hình đã lưu tại {MODEL_PATH}")
    return model


# =============================
# 🚀 Chạy huấn luyện
# =============================
if __name__ == "__main__":
    X, y, scaler = load_and_preprocess()
    train_gru_model(X, y, continue_training=True)

# NeuralOps Mobile

Expo React Native companion app for on-call SRE workflows.

## Features

- Dashboard usage overview
- Firing alerts list
- Active incidents list

## Setup

```bash
cd mobile
npm install
export EXPO_PUBLIC_API_BASE=http://localhost:8080/api/v1
export EXPO_PUBLIC_API_KEY=demo-api-key
npm start
```

Scan the QR code with Expo Go (iOS/Android) or run `npm run android` / `npm run ios`.

## Configuration

| Variable | Default |
|----------|---------|
| `EXPO_PUBLIC_API_BASE` | `http://localhost:8080/api/v1` |
| `EXPO_PUBLIC_API_KEY` | `demo-api-key` |
| `EXPO_PUBLIC_TENANT_ID` | demo tenant UUID |

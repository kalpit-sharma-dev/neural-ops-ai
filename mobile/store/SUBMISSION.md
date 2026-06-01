# App Store / Play Store submission

## Prerequisites

1. Expo account + EAS CLI: `npm i -g eas-cli && eas login`
2. Apple Developer Program + Google Play Console access
3. Replace placeholders in `eas.json`:
   - `ascAppId` — App Store Connect app ID
   - `appleTeamId` — Apple team ID
   - `google-play-service-account.json` — Play Console service account

## Build

```bash
cd mobile
eas build --platform all --profile production
```

## Submit

```bash
eas submit --platform ios --profile production
eas submit --platform android --profile production
```

## SDK embedding

Third-party apps can depend on `@neuralops/mobile-sdk`:

```typescript
import { configureMobileSDK, trackScreen } from '@neuralops/mobile-sdk';

configureMobileSDK({
  apiBase: 'https://api.neuralops.ai/api/v1',
  tenantId: 'your-tenant-id',
});
trackScreen('Home');
```

Bundle IDs: `ai.neuralops.mobile` (iOS + Android) — registered in `app.json`.

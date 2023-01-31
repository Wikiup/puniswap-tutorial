import React from 'react';
import { BrowserRouter } from 'react-router-dom';
import { createRoot } from 'react-dom/client';

import App from '@/App';
import '@/assets/scss/style.scss';
import IncognitoWalletProvider from '@/hooks/useIncognito';

const container = document.getElementById('root');
const root = createRoot(container!);
const app = (
  <BrowserRouter>
    <IncognitoWalletProvider>
      <App />
    </IncognitoWalletProvider>
  </BrowserRouter>
);
root.render(app);

import { AccountDetail, AccountInfo, WalletState } from '@/types/incognito';
import React, { useContext, useState } from 'react';

interface IncognitoWalletContextType {
  isIncognitoInstalled: () => boolean;
  getWalletState: () => Promise<WalletState | undefined>;
  requestIncognitoAccount: () => Promise<AccountInfo | undefined>;
  requestSignTransaction: (payload: any) => any;
  showPopup: () => void;
  getIncognitoInject: () => void;
  currentAccount: AccountDetail | undefined;
  walletState: WalletState | undefined;
}

// @ts-ignore
const getIncognitoInject = () => window.incognito;

const IncognitoWalletContext = React.createContext<IncognitoWalletContextType>({
  isIncognitoInstalled: () => false,
  getWalletState: async () => undefined,
  requestIncognitoAccount: async () => undefined,
  requestSignTransaction: () => null,
  showPopup: () => null,
  getIncognitoInject: () => null,
  currentAccount: undefined,
  walletState: undefined,
});

const IncognitoWalletProvider = (props: any) => {
  const children = React.useMemo(() => props.children, []);
  const [walletState, setWalletState] = useState<WalletState | undefined>(undefined);
  const [account, setAccount] = useState<AccountInfo | undefined>(undefined);
  const [currentAccount, setCurrentAccount] = useState<AccountDetail | undefined>(undefined);

  const isIncognitoInstalled = (): boolean => {
    // @ts-ignore
    return typeof window.incognito !== 'undefined';
  };

  const getWalletState = async (): Promise<WalletState | undefined> => {
    let state = undefined;
    const incognito = getIncognitoInject();
    try {
      if (!incognito) return;
      const { result }: { result: { state: WalletState } } = await incognito.request({
        method: 'wallet_getState',
        params: {},
      });
      state = result.state;
      setWalletState(state);
      console.log('INCOGNITO getWalletState: ', state);
    } catch (e) {
      console.log('REQUEST GET WALLET STATE ERROR', e);
    }
    return state;
  };

  const requestIncognitoAccount = async (): Promise<AccountInfo | undefined> => {
    let account = undefined;
    const incognito = getIncognitoInject();
    try {
      if (!incognito) return;
      const state = (await getWalletState()) || {};
      if (state === 'unlocked') {
        const { result } = await incognito.request({
          method: 'wallet_requestAccounts',
          params: {},
        });
        if (result) {
          account = result;
        }
      }
    } catch (e) {
      console.log('REQUEST INCOGNITO ACCOUNT', e);
    }
    setAccount(account);
    if (account && account.accounts) {
      setCurrentAccount(account.accounts[0]);
    }
    return account;
  };

  const requestSignTransaction = async (payload: any) => {
    const incognito = getIncognitoInject();
    try {
      if (!incognito) return;
      const { result }: { result: { state: WalletState } } = await incognito.request({
        method: 'wallet_signTransaction',
        params: {
          ...payload,
        },
      });
      return Promise.resolve(result);
    } catch (e) {
      return Promise.reject(e);
    }
  };

  const showPopup = () => {
    const incognito = getIncognitoInject();
    try {
      if (incognito) {
        incognito.request({
          method: 'wallet_showPopup',
          params: {},
        });
      } else {
        window.open('https://chrome.google.com/webstore/detail/incognito-wallet/chngojfpcfnjfnaeddcmngfbbdpcdjaj');
      }
    } catch (e) {
      console.log('SHOW POPUP WITH ERROR: ', e);
    }
  };

  const listenerExtensionEvents = () => {
    const incognito = getIncognitoInject();
    if (incognito) {
      // Listener event lock | unlock
      incognito.on('stateChanged', async (result: any) => {
        if (result) {
          if (result.state === 'unlocked') {
            requestIncognitoAccount().then();
            setWalletState('unlocked');
          }
          if (result.state === 'locked') {
            setWalletState('locked');
            setAccount(undefined);
            setCurrentAccount(undefined);
          }
        }
      });

      // listen event change account
      incognito.on('accountsChanged', () => {
        requestIncognitoAccount().then();
      });
    }
  };

  React.useEffect(() => {
    listenerExtensionEvents();
    const getInfo = async () => {
      const incognito = getIncognitoInject();
      if (!incognito) return;
      const state = await getWalletState();
      if (state === 'unlocked') {
        requestIncognitoAccount().then();
      }
    };
    getInfo().then(() => {
      setTimeout(() => getInfo(), 4000);
    });
  }, []);

  return (
    <IncognitoWalletContext.Provider
      value={{
        isIncognitoInstalled,
        getWalletState,
        requestIncognitoAccount,
        requestSignTransaction,
        showPopup,
        getIncognitoInject,
        currentAccount,
        walletState,
      }}
    >
      <>{children}</>
    </IncognitoWalletContext.Provider>
  );
};

export const useIncognitoWallet = (): IncognitoWalletContextType => {
  const context = useContext(IncognitoWalletContext);
  if (!context) {
    throw new Error(
      'Incognito wallet context not found, useModal must be used within the <IncognitoWalletProvider>..</IncognitoWalletProvider>',
    );
  }
  return context;
};

export default IncognitoWalletProvider;

export { IncognitoWalletContext };

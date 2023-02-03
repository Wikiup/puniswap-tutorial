type WalletState = 'uninitialized' | 'locked' | 'unlocked';

type Balance = {
  amount: string;
  id: string; // tokenID
};

type AccountDetail = {
  keyDefine: string; // account ID
  balances: Balance[];
  paymentAddress: string; // incognito address
};

type AccountInfo = {
  accounts: AccountDetail[];
  otaReceiver: string;
};

export { WalletState, AccountDetail, AccountInfo, Balance };

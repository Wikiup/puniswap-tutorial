import React, { useState } from 'react';
import Button from '@mui/material/Button';
import { useIncognitoWallet } from '@/hooks/useIncognito';
import { shortenString } from '@/utils/string';
import Box from '@mui/material/Box';
import { Container } from '@mui/material';
import Typography from '@mui/material/Typography';
import { Token } from '@/types/token';
import { Balance } from '@/types/incognito';
import BigNumber from 'bignumber.js';
import incSvg from '@/images/incognito.svg';

const Main = () => {
  const { requestIncognitoAccount, showPopup, getWalletState, requestSignTransaction, currentAccount, walletState } =
    useIncognitoWallet();

  const [showAccount, setShowAccount] = useState<boolean>(false);
  const [tokens, setTokens] = useState<Token[] | undefined>(undefined);

  const handleConnect = async () => {
    const state = await getWalletState();
    switch (state) {
      // show popup enter password
      case 'locked':
        return showPopup();
      // get current account info
      case 'unlocked':
        return requestIncognitoAccount().then();
      default:
        // install wallet
        window.open('https://chrome.google.com/webstore/detail/incognito-wallet/chngojfpcfnjfnaeddcmngfbbdpcdjaj');
    }
  };

  const handleShowAccount = async () => {
    if (!walletState || walletState === 'locked') return setShowAccount(false);
    const account = await requestIncognitoAccount();
    if (account && account.accounts) {
      setShowAccount(true);
    }
  };

  const handleSignTransaction = async () => {
    if (!walletState || walletState === 'locked') return;
    // A sample payload of swapping pUSDT (unified) for pMatic via Uniswap in Ethereum network
    const burnAmount = 1000000;
    const externalNetworkID = 1; // ETH = 1

    const USDT_UNIFIED_TOKENID = '076a4423fa20922526bd50b0d7b0dc1c593ce16e15ba141ede5fb5a28aa3f229';
    const token = findToken(USDT_UNIFIED_TOKENID);
    if (!token) return alert(`Cant find token ${USDT_UNIFIED_TOKENID}`);
    const incToken = token.ListUnifiedToken.length
      ? token.ListUnifiedToken.find(token => token.NetworkID === externalNetworkID)
      : token;

    if (!incToken) return alert(`Cant find incToken ${USDT_UNIFIED_TOKENID} network ${externalNetworkID}`);

    const incTokenID = incToken.TokenID;

    const account = await requestIncognitoAccount();

    if (!account) return alert(`Cant get account`);

    const payload = {
      metadata: {
        Data: [
          {
            IncTokenID: incTokenID, // USDT TokenID with Ethereum Network
            RedepositReceiver: account.otaReceiver, // OTA receiver via requestIncognitoAccount field otaReceiver
            BurningAmount: `${burnAmount}`,
            ExternalNetworkID: externalNetworkID, // Ethereum NetworkID is 3
            ExternalCalldata:
              '421f4388000000000000000000000000c2132d05d31c914a87c6611c10748aeb04b58e8f0000000000000000000000000d500b1d8e8ef31e21c99d1db9a6444d3adf127000000000000000000000000000000000000000000000000000000000000001f4000000000000000000000000cc8c88e9dae72fa07ac077933a2e73d146fecdf000000000000000000000000000000000000000000000000000000000000003e7000000000000000000000000000000000000000000000000000311832a534fa700000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001',
            ExternalCallAddress: 'CC8c88e9Dae72fa07aC077933a2E73d146FECdf0', // removed prefix 0x
            ReceiveToken: '0000000000000000000000000000000000000000', // Contact address receive token
            WithdrawAddress: '0000000000000000000000000000000000000000',
          },
        ],
        BurnTokenID: USDT_UNIFIED_TOKENID, // USDT(unified)
        Type: 348,
      },
      info: '',
      networkFee: 100000000,
      prvPayments: [],
      tokenPayments: [
        {
          // fee receiver address
          PaymentAddress:
            '12stZ9UxpNNd8oKjuf5Bpfb44AvrogVVZCVjjF2u9K2Q1s55LHrQxXypSRZ9BV1PtphRf1JxiBaKbmhmKdj3c7DWt4kkcKV4HyWcuws8YPiZbHuDxFxdar9vQvbB3pvYGQAaj4PR38Sr52fiTDPn',
          // fee estimated
          Amount: '33593747',
          Message: '',
        },
        {
          // burn address hardcode
          PaymentAddress:
            '12RxahVABnAVCGP3LGwCn8jkQxgw7z1x14wztHzn455TTVpi1wBq9YGwkRMQg3J4e657AbAnCvYCJSdA9czBUNuCKwGSRQt55Xwz8WA',
          Amount: burnAmount,
          Message: '',
        },
      ],
      tokenID: USDT_UNIFIED_TOKENID,
      txType: 7,
      receiverAddress:
        '12su5sVeYCEjQfcn5XCzkKsSmctXrmhvmPmc5C4gDGk8hAAB4upSQqoFB7zVijw7gsvnUyjdx5m6GzKDqSps2v2gjxTxmErq48wqWU4qDsG61HrERu6odcZ5dAztD96Vtxy3XeYSrkWFEshfe8wZ',
      isSignAndSendTransaction: false,
    };

    const tx = await requestSignTransaction(payload);
    console.log('tx: ', tx); // { txhash, rawData }
  };

  const fetchTokens = () => {
    fetch('https://api-webapp.incognito.org/tokenlist')
      .then(res => res.json())
      .then(result => {
        if (result.Result) {
          const tokens: Token[] = result.Result;
          setTokens(tokens);
        }
      });
  };

  const findToken = (tokenID: string) => {
    return tokens?.find(token => token.TokenID === tokenID);
  };

  const renderContent = (item: Balance) => {
    const tokenID = item.id;
    const token = findToken(tokenID);
    return (
      <Box display="flex" sx={{ alignItems: 'center', marginTop: '18px' }}>
        <Box display="flex" sx={{ alignItems: 'center', flex: 1 }}>
          <Typography className="symbol">
            {token?.Symbol}_{token?.Network}
          </Typography>
        </Box>
        <Typography className="amount">
          {new BigNumber(item.amount).div(new BigNumber(10).pow(token?.PDecimals || 0)).toString()}&nbsp;
          {token?.Symbol}
        </Typography>
      </Box>
    );
  };

  React.useEffect(() => {
    fetchTokens();
  }, []);

  return (
    <Container
      sx={{
        paddingTop: '40px',
        paddingBottom: '40px',
        paddingLeft: '120px',
        paddingRight: '120px',
        backgroundColor: 'transparent',

        margin: '0',
        marginTop: '20px',
        position: 'absolute',
        top: '30%',
        left: '50%',
        transform: 'translate(-50%, -50%)',

        display: 'flex',
        flexDirection: 'row',
        justifyContent: 'center',
      }}
      style={{ maxWidth: '1100px' }}
    >
      <Box
        sx={{
          width: 400,
          backgroundColor: '#252525',
          borderRadius: '4px',
          padding: '16px',
          boxShadow: '0 2px 5px 0 rgb(0 0 0 / 30%), 0 2px 10px 0 rgb(0 0 0 / 30%)',
          marginRight: '40px',
        }}
      >
        <Typography className="typography">Basic Actions</Typography>
        <Button className="button" variant="contained" onClick={handleConnect} style={{ marginTop: 24 }}>
          {currentAccount ? shortenString(currentAccount.paymentAddress) : walletState ? 'CONNECT' : 'INSTALL WALLET'}
        </Button>
        <Button className="button" variant="contained" onClick={handleShowAccount} style={{ marginTop: 24 }}>
          INCOGNITO ACCOUNT
        </Button>
        <Button
          className="button"
          variant="contained"
          onClick={handleSignTransaction}
          style={{ marginTop: 24, marginBottom: 24 }}
        >
          SIGN TRANSACTION
        </Button>
      </Box>
      {showAccount && currentAccount && currentAccount.balances && (
        <Box
          sx={{
            flex: 1,
            backgroundColor: '#252525',
            borderRadius: '4px',
            padding: '16px',
            boxShadow: '0 2px 5px 0 rgb(0 0 0 / 30%), 0 2px 10px 0 rgb(0 0 0 / 30%)',
            marginRight: '40px',
          }}
        >
          {currentAccount.balances.map(renderContent)}
        </Box>
      )}
    </Container>
  );
};

export default Main;

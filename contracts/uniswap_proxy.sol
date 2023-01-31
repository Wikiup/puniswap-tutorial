// SPDX-License-Identifier: MIT
pragma solidity 0.8.17;

import './utils/trade_utils.sol';
import './interfaces/IERC20.sol';

interface UniswapV2 {
    function swapExactTokensForTokens(
        uint amountIn,
        uint amountOutMin,
        address[] calldata path,
        address to,
        uint deadline
    ) external returns (uint[] memory amounts);
    function swapExactETHForTokens(uint amountOutMin, address[] calldata path, address to, uint deadline)
    external
    payable
    returns (uint[] memory amounts);
    function swapExactTokensForETH(uint amountIn, uint amountOutMin, address[] calldata path, address to, uint deadline)
    external
    returns (uint[] memory amounts);
}

contract UniswapV2Trade is TradeUtils {
    // Variables
    UniswapV2 constant public uniswapV2 = UniswapV2(0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D);
    address constant public wETH = 0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2; // mainnet address

    // Receive function which allows contract receive native token.
    receive() external payable {}

    function trade(address[] memory path, uint amountOutMin) public payable returns (address, uint) {
        require(path.length >= 2, "Proxy: invalid path");
        uint256 swapAmount = msg.value > 0 ? msg.value : balanceOf(IERC20(path[0]));
        uint[] memory amounts;
        bool isSwapForNative = false;
        if (msg.value == 0) {
            // approve
            approve(IERC20(path[0]), address(uniswapV2), swapAmount);
            if (path[path.length - 1] != wETH) { // token to token.
                amounts = tokenToToken(path, swapAmount, amountOutMin);
            } else {
                amounts = tokenToEth(path, swapAmount, amountOutMin);
                isSwapForNative = true;
            }
        } else {
            amounts = ethToToken(path, swapAmount, amountOutMin);
        }
        require(amounts.length >= 2, "Proxy: invalid response values");
        require(amounts[amounts.length - 1] >= amountOutMin && amounts[0] == swapAmount);
        // ETH_CONTRACT_ADDRESS is a address present for eth native
        return (isSwapForNative ? address(ETH_CONTRACT_ADDRESS) : path[path.length - 1], amounts[amounts.length - 1]);
    }

    function ethToToken(address[] memory path, uint srcQty, uint amountOutMin) internal returns (uint[] memory) {
        return uniswapV2.swapExactETHForTokens{value: srcQty}(amountOutMin, path, msg.sender, block.timestamp + 1000000);
    }

    function tokenToEth(address[] memory path, uint srcQty, uint amountOutMin) internal returns (uint[] memory) {
        return uniswapV2.swapExactTokensForETH(srcQty, amountOutMin, path, msg.sender, block.timestamp + 1000000);
    }

    function tokenToToken(address[] memory path, uint srcQty, uint amountOutMin) internal returns (uint[] memory) {
        return uniswapV2.swapExactTokensForTokens(srcQty, amountOutMin, path, msg.sender, block.timestamp + 1000000);
    }
}

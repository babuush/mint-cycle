// SPDX-License-Identifier: GPL-3.0-only
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC721/ERC721.sol";

contract ProductNFT is ERC721 {
    constructor() ERC721("RetailProduct", "PROD") {}

    // Custom Mint function that accepts a specific ID
    // This allows the Central Database to dictate IDs (Timestamp-based)
    // preventing "Split Brain" issues between DB and Blockchain.
    function mint(address to, uint256 id) public {
        _mint(to, id);
    }
}

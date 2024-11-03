// convertToBech32.js
const { bech32 } = require("bech32");

// Strip whitespace and validate the input
const inputAddress = process.argv[2].trim();

try {
    // Decode the input Cosmos address from Bech32
    const decoded = bech32.decode(inputAddress);
    
    // Re-encode with "arkeo" prefix using the same data (words)
    const arkeoAddress = bech32.encode("arkeo", decoded.words);
    console.log(arkeoAddress);
} catch (error) {
    console.error(`Error converting address "${inputAddress}":`, error.message);
    process.exit(1);
}

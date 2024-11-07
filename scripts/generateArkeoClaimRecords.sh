#!/bin/bash

# Input and output files
INPUT_FILE="arkeo_airdrop.csv"
OUTPUT_FILE="arkeo_airdrop_bech32.csv"
TEMP_FILE="/tmp/converted_arke_airdrop.csv"
SORTED_TEMP_FILE="/tmp/sorted_arke_airdrop.csv"

# Initialize the output CSV with header
echo "address,distributed_amount" >"$TEMP_FILE"

# Convert each address to Bech32 format with "arkeo" prefix and aggregate the distributed amount
awk -F',' 'NR > 1 { print $1 "," $3 }' "$INPUT_FILE" | while IFS=, read -r original_address distributed_amount; do
    original_address=$(echo "$original_address" | xargs)

    arkeo_address=$(node convertToBech32.js "$original_address")

    if [ -n "$arkeo_address" ]; then
        echo "$arkeo_address,$distributed_amount" >>"$TEMP_FILE"
    else
        echo "Conversion failed for address: $original_address"
    fi
done

# Aggregate `distributed_amount` for each unique Bech32 address with "arkeo" prefix
awk -F',' '
    NR == 1 { next } # Skip header row
    {
        address = $1
        distributed_sum[address] += $2
    }
    END {
        for (addr in distributed_sum) {
            if (distributed_sum[addr] > 0) {
                printf "%s,%.2f\n", addr, distributed_sum[addr]
            } 
        }
    }
' "$TEMP_FILE" >"$SORTED_TEMP_FILE"

# Sort by `distributed_amount` in descending order and add header to final output
echo "address,distributed_amount" > "$OUTPUT_FILE"
sort -t, -k2,2nr "$SORTED_TEMP_FILE" >> "$OUTPUT_FILE"

# Clean up temporary files
rm "$TEMP_FILE" "$SORTED_TEMP_FILE"

echo "Converted, aggregated, and sorted CSV has been saved to $OUTPUT_FILE"

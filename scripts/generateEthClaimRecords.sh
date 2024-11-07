#!/bin/bash

CSV_FILES=(
    "../csv/airdrop_ETH_FOX.csv" 
    "../csv/airdrop_ETH_FOXy.csv" 
    "../csv/airdrop_ETH_scFOX0324.csv" 
    "../csv/airdrop_ETH_tFOX.csv" 
    "../csv/airdrop_GNO_FOX.csv" 
    "../csv/airdrop_POLY_FOX.csv" 
    "../csv/airdrop_farm_ETH_HEDGEY.csv" 
    "../csv/airdrop_farm_GNO_HEDGEY.csv"
)

CSV_FILES_WITH_MULTIPLIER=(
    "../csv/airdrop_ETH_UNI-V2.csv"
    "../csv/airdrop_farm_ETH_STAKINGREWARDSv5.csv"
    "../csv/airdrop_farm_ETH_STAKINGREWARDSv6.csv"
    "../csv/airdrop_farm_ETH_STAKINGREWARDSv7.csv"
    "../csv/airdrop_farm_ETH_STAKINGREWARDSv8.csv"
)

MULTIPLIER=460  # The multiplier to apply to the second set of CSV files
TOTAL_DISTRIBUTION=12705000  # Total amount to distribute
OUTPUT_FILE="eth_airdrop.csv"
TEMP_FILE="/tmp/combined.csv"
TEMP_WITH_DISTRIBUTION="/tmp/combined_with_distribution.csv"

# Initialize output CSV with a header
echo "address,time_weighted_average" > "$TEMP_FILE"

# Process initial CSV files (no multiplier) and aggregate by address (case-insensitive)
declare -A totals
declare -A original_case  # Store the original case-sensitive address

for CSV in "${CSV_FILES[@]}"; do
    CSV_PATH="$(pwd)/$CSV"

    if [ ! -f "$CSV_PATH" ]; then
      echo "File not found: $CSV_PATH"
      continue
    fi

    # Accumulate time-weighted averages without multiplier, case-insensitively
    awk -F',' 'NR > 1 {
        address_lc = tolower($1)
        totals[address_lc] += $2
        if (!(address_lc in original_case)) {
            original_case[address_lc] = $1  # Store original case if first occurrence
        }
    } END {
        for (address in totals) {
            printf "%s,%.18f\n", original_case[address], totals[address]
        }
    }' "$CSV_PATH" >> "$TEMP_FILE"
done

# Process CSV files with the multiplier and aggregate by address (case-insensitive)
for CSV in "${CSV_FILES_WITH_MULTIPLIER[@]}"; do
    CSV_PATH="$(pwd)/$CSV"

    if [ ! -f "$CSV_PATH" ]; then
      echo "File not found: $CSV_PATH"
      continue
    fi

    # Accumulate time-weighted averages with multiplier applied, case-insensitively
    awk -F',' -v multiplier="$MULTIPLIER" '
        NR > 1 {
            address_lc = tolower($1)
            totals[address_lc] += $2 * multiplier
            if (!(address_lc in original_case)) {
                original_case[address_lc] = $1  # Store original case if first occurrence
            }
        }
        END {
            for (address in totals) {
                printf "%s,%.18f\n", original_case[address], totals[address]
            }
        }
    ' "$CSV_PATH" >> "$TEMP_FILE"
done

# Summarize and calculate total for distribution
total_weighted_sum=$(awk -F',' '
    NR > 1 { sum += $2 } # Sum all time_weighted_average values
    END { print sum }
' "$TEMP_FILE")

# Calculate and distribute the final amount for each address, using unique case-insensitive addresses
awk -F',' -v total_weighted_sum="$total_weighted_sum" -v total_distribution="$TOTAL_DISTRIBUTION" '
    NR == 1 { next } # Skip header row
    {
        address = $1
        weighted_sum[address] += $2
    }
    END {
        for (addr in weighted_sum) {
            proportion = weighted_sum[addr] / total_weighted_sum
            distributed_amount = proportion * total_distribution
            if (distributed_amount > 0) { # Only include non-zero distributed amounts
                printf "%s,%.2f\n", addr, distributed_amount >> "'"$TEMP_WITH_DISTRIBUTION"'"
            }
        }
    }
' "$TEMP_FILE"

# Sort by distributed_amount in descending order and add header to final output
echo "address,distributed_amount" > "$OUTPUT_FILE"
sort -t, -k2,2nr "$TEMP_WITH_DISTRIBUTION" >> "$OUTPUT_FILE"

# Clean up temporary files
rm "$TEMP_FILE" "$TEMP_WITH_DISTRIBUTION"

echo "Combined, distributed, and sorted CSV has been saved to $OUTPUT_FILE"

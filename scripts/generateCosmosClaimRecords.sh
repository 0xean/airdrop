#!/bin/bash

# Verify Bash is being used
if [ -z "$BASH_VERSION" ]; then
    echo "This script requires bash. Please run it with bash."
    exit 1
fi

#1.21m for testnet

# List of CSV files
CSV_FILES=(
    "../csv/airdrop_GAIA_delegates.csv"
    "../csv/airdrop_OSMO_delegates.csv"
    "../csv/airdrop_THOR_bonders.csv"
    "../csv/airdrop_THOR_liquidity.csv"
    "../csv/airdrop_OSMO_LP.csv"
)

# Corresponding distribution amounts for each file
DISTRIBUTION_AMOUNTS=(3781250 3781250 3000000 1890625 1890625)

# Corresponding distribution types for each file (either "even" or "weighted")
DISTRIBUTION_TYPES=("weighted" "weighted" "even" "even" "even")

OUTPUT_FILE="arkeo_airdrop.csv"
TEMP_DIR="/tmp/airdrop_calculations"
mkdir -p "$TEMP_DIR"

# Initialize final output with a header
echo "address,time_weighted_average,distributed_amount" > "$OUTPUT_FILE"

# Loop over each file and process with its respective distribution amount and type
for i in "${!CSV_FILES[@]}"; do
    CSV="${CSV_FILES[i]}"
    TOTAL_DISTRIBUTION="${DISTRIBUTION_AMOUNTS[i]}"
    DISTRIBUTION_TYPE="${DISTRIBUTION_TYPES[i]}"
    TEMP_FILE="$TEMP_DIR/$(basename "$CSV" .csv)_processed.csv"

    if [ ! -f "$CSV" ]; then
      echo "File not found: $CSV"
      continue
    fi

    # Step 1: Aggregate and calculate distributed amount based on type
    if [[ "$DISTRIBUTION_TYPE" == "weighted" ]]; then
        # Proportional distribution based on time-weighted averages
        awk -F',' -v OFS=',' -v total_distribution="$TOTAL_DISTRIBUTION" '
            NR > 1 {
                address_lc = tolower($1)
                totals[address_lc] += $2
                if (!(address_lc in original_case)) {
                    original_case[address_lc] = $1  # Store original case if first occurrence
                }
            }
            END {
                total_weighted_sum = 0
                for (address in totals) {
                    total_weighted_sum += totals[address]
                }
                for (address in totals) {
                    proportion = totals[address] / total_weighted_sum
                    distributed_amount = proportion * total_distribution
                    printf "%s,%.18f,%.2f\n", original_case[address], totals[address], distributed_amount
                }
            }
        ' "$CSV" > "$TEMP_FILE"
    elif [[ "$DISTRIBUTION_TYPE" == "even" ]]; then
        # Even distribution across all addresses
        awk -F',' -v OFS=',' -v total_distribution="$TOTAL_DISTRIBUTION" '
            NR > 1 {
                address_lc = tolower($1)
                totals[address_lc] += $2
                if (!(address_lc in original_case)) {
                    original_case[address_lc] = $1  # Store original case if first occurrence
                }
                count++
            }
            END {
                even_share = total_distribution / count
                for (address in totals) {
                    printf "%s,%.18f,%.2f\n", original_case[address], totals[address], even_share
                }
            }
        ' "$CSV" > "$TEMP_FILE"
    fi

    # Step 2: Append the processed results to the final output file
    cat "$TEMP_FILE" >> "$OUTPUT_FILE"
done

# Step 3: Aggregate across all files, sort, and remove duplicates by address
awk -F',' '
    NR == 1 { next } # Skip header row
    {
        address_lc = tolower($1)
        if (!(address_lc in original_case)) {
            original_case[address_lc] = $1  # Preserve original case of first occurrence
        }
        aggregated_weighted[address_lc] += $2
        aggregated_distributed[address_lc] += $3
    }
    END {
        for (address in aggregated_weighted) {
            printf "%s,%.18f,%.2f\n", original_case[address], aggregated_weighted[address], aggregated_distributed[address]
        }
    }
' "$OUTPUT_FILE" | sort -t, -k3,3nr > "$TEMP_DIR/final_sorted.csv"

# Add header and save final sorted and aggregated output
echo "address,time_weighted_average,distributed_amount" > "$OUTPUT_FILE"
cat "$TEMP_DIR/final_sorted.csv" >> "$OUTPUT_FILE"

# Clean up temporary files
rm -rf "$TEMP_DIR"

echo "Final combined and sorted CSV has been saved to $OUTPUT_FILE"

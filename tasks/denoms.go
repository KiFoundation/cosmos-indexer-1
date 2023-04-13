package tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DefiantLabs/cosmos-tax-cli/config"
	dbTypes "github.com/DefiantLabs/cosmos-tax-cli/db"
	"github.com/DefiantLabs/cosmos-tax-cli/juno"
	"github.com/DefiantLabs/cosmos-tax-cli/osmosis"
	"github.com/DefiantLabs/cosmos-tax-cli/rest"

	"gorm.io/gorm"
)

type OsmosisAssets struct {
	Assets []Asset
}

type AssetList struct {
	Assets []Asset
}

type Asset struct {
	Denoms []DenomUnit `json:"denom_units"`
	Symbol string
	Base   string
	Name   string
}

type DenomUnit struct {
	Denom    string
	Exponent int
	Aliases  []string
}

func DoChainSpecificUpsertDenoms(db *gorm.DB, chain string) {
	if chain == osmosis.ChainID {
		config.Log.Info("Updating Omsosis specific denoms")
		UpsertOsmosisDenoms(db)
	}

	if chain == juno.ChainID {
		config.Log.Info("Updating Juno specific denoms")
		UpsertJunoDenoms(db)
	}
	// may want to move this elsewhere, or eliminate entirely
	// I would prefer we just grab the denoms when needed always
	// Current problem: we use the denom cache in various blocks later on
	dbTypes.CacheDenoms(db)
}

func UpsertOsmosisDenoms(db *gorm.DB) {
	url := "https://raw.githubusercontent.com/osmosis-labs/assetlists/main/osmosis-1/osmosis-1.assetlist.json"

	denomAssets, err := getAssetsList(url)
	if err != nil {
		config.Log.Fatal("Download Osmosis Denom Metadata", err)
	} else {
		denoms := assetListToDenoms(denomAssets)
		err = dbTypes.UpsertDenoms(db, denoms)
		if err != nil {
			config.Log.Fatal("Upsert Osmosis Denom Metadata", err)
		}
	}
}

func UpsertJunoDenoms(db *gorm.DB) {
	url := "https://raw.githubusercontent.com/cosmos/chain-registry/master/juno/assetlist.json"

	denomAssets, err := getAssetsList(url)
	if err != nil {
		config.Log.Fatal("Error downloading Juno Denom Metadata", err)
	} else {
		denoms := assetListToDenoms(denomAssets)
		err = dbTypes.UpsertDenoms(db, denoms)
		if err != nil {
			config.Log.Fatal("Error upserting Juno Denom Metadata", err)
		}
	}
}

func assetListToDenoms(assets *AssetList) []dbTypes.DenomDBWrapper {
	var denoms []dbTypes.DenomDBWrapper = make([]dbTypes.DenomDBWrapper, len(assets.Assets))
	for i, asset := range assets.Assets {
		denoms[i].Denom = dbTypes.Denom{Base: asset.Base, Name: asset.Name, Symbol: asset.Symbol}
		denoms[i].DenomUnits = make([]dbTypes.DenomUnitDBWrapper, len(asset.Denoms))

		for ii, denomUnit := range asset.Denoms {
			denoms[i].DenomUnits[ii].DenomUnit = dbTypes.DenomUnit{Exponent: uint(denomUnit.Exponent), Name: denomUnit.Denom}
			denoms[i].DenomUnits[ii].Aliases = make([]dbTypes.DenomUnitAlias, len(denomUnit.Aliases))

			for iii, denomUnitAlias := range denomUnit.Aliases {
				denoms[i].DenomUnits[ii].Aliases[iii] = dbTypes.DenomUnitAlias{Alias: denomUnitAlias}
			}
		}
	}

	return denoms
}

func getAssetsList(assetsURL string) (*AssetList, error) {
	assets := &AssetList{}
	err := getJSON(assetsURL, assets)
	if err != nil {
		return nil, err
	}

	return assets, nil
}

func getJSON(url string, target interface{}) error {
	var myClient = &http.Client{Timeout: 10 * time.Second}

	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("got status code: %v from url: %v", r.Status, url)
	}

	return json.NewDecoder(r.Body).Decode(target)
}

func DenomUpsertTask(apiHost string, db *gorm.DB) {
	config.Log.Info(fmt.Sprintf("Updating Denom Metadata from %s", apiHost))
	denomsMetadata, err := rest.GetDenomsMetadatas(apiHost)
	if err != nil {
		config.Log.Error(fmt.Sprintf("Error in Denom Metadata Update task when reaching out to the API at %s ", apiHost), err)
		return
	}

	var denoms []dbTypes.DenomDBWrapper = make([]dbTypes.DenomDBWrapper, len(denomsMetadata.Metadatas))
	for i, denom := range denomsMetadata.Metadatas {
		if denom.Name == "" {
			denom.Name = "UNKNOWN"
		}
		if denom.Symbol == "" {
			denom.Symbol = "UNKNOWN"
		}

		denoms[i].Denom = dbTypes.Denom{Base: denom.Base, Name: denom.Name, Symbol: denom.Symbol}

		denoms[i].DenomUnits = make([]dbTypes.DenomUnitDBWrapper, len(denom.DenomUnits))

		for ii, denomUnit := range denom.DenomUnits {
			denoms[i].DenomUnits[ii].DenomUnit = dbTypes.DenomUnit{Exponent: uint(denomUnit.Exponent), Name: denomUnit.Denom}

			denoms[i].DenomUnits[ii].Aliases = make([]dbTypes.DenomUnitAlias, len(denomUnit.Aliases))

			for iii, denomUnitAlias := range denomUnit.Aliases {
				denoms[i].DenomUnits[ii].Aliases[iii] = dbTypes.DenomUnitAlias{Alias: denomUnitAlias}
			}
		}
	}

	err = dbTypes.UpsertDenoms(db, denoms)
	if err != nil {
		config.Log.Error("Error updating database in Denom Metadata Update task", err)
		return
	}
	config.Log.Info("Denom Metadata Update Complete")
}

func IBCDenomUpsertTask(apiHost string, db *gorm.DB) {
	config.Log.Info(fmt.Sprintf("Updating IBC Denom Metadata from %s", apiHost))
	ibcDenomsMetadata, err := rest.GetIBCDenomTraces(apiHost)
	if err != nil {
		config.Log.Error(fmt.Sprintf("Error in IBC Denom Metadata Update task when reaching out to the API at %s ", apiHost), err)
		return
	}

	denoms := make([]dbTypes.IBCDenom, len(ibcDenomsMetadata.DenomTraces))
	for i, t := range ibcDenomsMetadata.DenomTraces {
		denoms[i] = dbTypes.IBCDenom{
			Hash:      t.IBCDenom(),
			Path:      t.Path,
			BaseDenom: t.BaseDenom,
		}
	}

	err = dbTypes.UpsertIBCDenoms(db, denoms)
	if err != nil {
		config.Log.Error("Error updating database in IBC Denom Metadata Update task", err)
		return
	}
	config.Log.Info("IBC Denom Metadata Update Complete")
}

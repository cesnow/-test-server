package handler

import (
	"google.golang.org/protobuf/types/known/wrapperspb"
	"hash/crc32"
	configurationpb "kiyudesign.com/cesnow/light-server/pkg/tproto/bffpb/configuration"
)

func (c *ConfigurationHandler) HelpGetCountriesList(in *configurationpb.THelpGetCountriesList) (*configurationpb.Help_CountriesList, error) {
	// 1. 依序建立各國的 Help_Country 物件
	var countries []*configurationpb.Help_Country

	// -- 台灣 --
	taiwan := &configurationpb.Help_Country{
		Hidden:      false,
		Iso2:        "TW",
		DefaultName: "Taiwan",
		CountryCodes: []*configurationpb.Help_CountryCode{
			{
				CountryCode: "886",
				Prefixes:    []string{"886"},
				Patterns:    []string{"+886-*"},
			},
		},
	}
	// 根據 lang_code 決定是否帶入 localized name
	switch in.LangCode {
	case "zh-TW", "zh", "zh_Hant":
		taiwan.Name = &wrapperspb.StringValue{Value: "台灣"}
	case "en", "en-US", "en_GB":
		taiwan.Name = &wrapperspb.StringValue{Value: "Taiwan"}
	}
	countries = append(countries, taiwan)

	// -- 美國 --
	us := &configurationpb.Help_Country{
		Hidden:      false,
		Iso2:        "US",
		DefaultName: "United States",
		CountryCodes: []*configurationpb.Help_CountryCode{
			{
				CountryCode: "1",
				Prefixes:    []string{"1"},
				Patterns:    []string{"+1-*"},
			},
		},
	}
	switch in.LangCode {
	case "zh-TW", "zh", "zh_Hant":
		us.Name = &wrapperspb.StringValue{Value: "美國"}
	case "en", "en-US", "en_GB":
		us.Name = &wrapperspb.StringValue{Value: "United States"}
	}
	countries = append(countries, us)

	// （可依需求再加入其他國家，直接複製上面模式即可）

	// 2. 計算 CRC32 hash 作為版本檢查
	var hashBuf []byte
	for _, hc := range countries {
		hashBuf = append(hashBuf, []byte(hc.Iso2+hc.DefaultName)...)
	}
	currentHash := int32(crc32.ChecksumIEEE(hashBuf))

	// 3. 如果前端傳入的 hash 相同，就只回 hash，不重傳 countries
	if in.Hash == currentHash {
		return &configurationpb.Help_CountriesList{
			Hash: currentHash,
		}, nil
	}

	// 4. 回傳完整清單及新的 hash
	return &configurationpb.Help_CountriesList{
		Countries: countries,
		Hash:      currentHash,
	}, nil
}

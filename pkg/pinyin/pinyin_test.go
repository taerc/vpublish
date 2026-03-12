package pinyin

import "testing"

func TestGenerateCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"无人机", "TYPE_WU_REN_JI"},
		{"地面站", "TYPE_DI_MIAN_ZHAN"},
		{"无人机V2", "TYPE_WU_REN_JI_V2"},
		{"地面站Pro", "TYPE_DI_MIAN_ZHAN_PRO"},
		{"APP 1.0版本", "TYPE_APP_1_0_BAN_BEN"},
		{"无人机-专业版V2.1", "TYPE_WU_REN_JI_ZHUAN_YE_BAN_V2_1"},
		{"Drone", "TYPE_DRONE"},
		{"123", "TYPE_123"},
		{"V2", "TYPE_V2"},
		{"版本2", "TYPE_BAN_BEN_2"},
		{"V2版本", "TYPE_V2_BAN_BEN"},
		{"APP客户端1.0", "TYPE_APP_KE_HU_DUAN_1_0"},
		{" 无人机 ", "TYPE_WU_REN_JI"},
		{"无人机--V2", "TYPE_WU_REN_JI_V2"},
		{"", "TYPE_UNNAMED"},
		{"  ", "TYPE_UNNAMED"},
		{"---", "TYPE_UNNAMED"},
		{"APP1.0Pro", "TYPE_APP1_0PRO"},
		{"低空监测系统V3.2.1", "TYPE_DI_KONG_JIAN_CE_XI_TONG_V3_2_1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GenerateCode(tt.input)
			if result != tt.expected {
				t.Errorf("GenerateCode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidCode(t *testing.T) {
	tests := []struct {
		code     string
		expected bool
	}{
		{"TYPE_WU_REN_JI", true},
		{"TYPE_WU_REN_JI_V2", true},
		{"TYPE_DI_MIAN_ZHAN_PRO", true},
		{"TYPE_APP1_0_BAN_BEN", true},
		{"TYPE_123", true},
		{"TYPE_V2", true},
		{"TYPE_", false},
		{"TYPE_a", false},
		{"TYPE_test", false},
		{"INVALID_CODE", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := IsValidCode(tt.code)
			if result != tt.expected {
				t.Errorf("IsValidCode(%q) = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

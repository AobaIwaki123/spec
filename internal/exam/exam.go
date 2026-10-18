package exam

import (
	"fmt"
	"strings"
)

// ID represents an IPA exam identifier.
type ID string

const (
	DB ID = "db" // データベーススペシャリスト (Database Specialist)
	NW ID = "nw" // ネットワークスペシャリスト (Network Specialist)
	SC ID = "sc" // 情報処理安全確保支援士 (Registered Information Security Specialist)
	PM ID = "pm" // プロジェクトマネージャ (Project Manager)
	SA ID = "sa" // システムアーキテクト (System Architect)
	ST ID = "st" // ITストラテジスト (IT Strategist)
	ES ID = "es" // エンベデッドシステムスペシャリスト (Embedded Systems Specialist)
	AU ID = "au" // システム監査技術者 (Systems Auditor)
)

// Definition holds all metadata, syllabus categories, scoring weights, and IPA URL patterns for an exam.
type Definition struct {
	ID              ID
	Name            string
	ShortName       string
	OfficialTitle   string
	PassingScore    int
	CategoryWeights map[string]float64
	SyllabusTopics  []string
	IPACode         string
	URLPrefix       string
	SupportedYears  []int
}

// Registry provides the catalog of supported IPA specialist exams.
var Registry = map[ID]Definition{
	DB: {
		ID:            DB,
		Name:          "データベーススペシャリスト",
		ShortName:     "DB",
		OfficialTitle: "高度情報処理技術者 データベーススペシャリスト試験",
		PassingScore:  60,
		CategoryWeights: map[string]float64{
			"正規化理論":            0.28,
			"トランザクション・同時実行制御": 0.24,
			"関係モデル・代数":        0.16,
			"障害回復":             0.12,
			"物理設計・インデックス":     0.12,
			"SQL・整合性制約":        0.08,
		},
		SyllabusTopics: []string{
			"データモデル（概念・論理・物理）",
			"正規化理論（第1正規形〜ボイス・コッド正規形）",
			"関係代数演算（選択・射影・結合・商・集合演算）",
			"トランザクション制御（ACID特性・直列化可能性・2相ロック）",
			"障害回復（WAL・チェックポイント・ロールフォワード/バック）",
			"物理DB設計（B+木インデックス・ハッシュ・パーティショニング）",
			"SQL最適化・実行計画・整合性制約",
		},
		IPACode:        "db",
		SupportedYears: []int{2024, 2023, 2022, 2021, 2020},
	},
	NW: {
		ID:            NW,
		Name:          "ネットワークスペシャリスト",
		ShortName:     "NW",
		OfficialTitle: "高度情報処理技術者 ネットワークスペシャリスト試験",
		PassingScore:  60,
		CategoryWeights: map[string]float64{
			"プロトコル・LAN/WAN技術": 0.30,
			"ルーティング・スイッチング":    0.25,
			"ネットワークセキュリティ":     0.25,
			"アプリケーションプロトコル":    0.10,
			"運用保守・トラフィック管理":    0.10,
		},
		SyllabusTopics: []string{
			"TCP/IPプロトコル群（IPv4/IPv6, TCP, UDP）",
			"ルーティングプロトコル（OSPF, BGP）",
			"レイヤ2/3スイッチング（VLAN, STP, VXLAN）",
			"ネットワークセキュリティ（TLS, IPsec, ゼロトラスト）",
			"DNS, DHCP, HTTP/3, QUIC",
		},
		IPACode:        "nw",
		SupportedYears: []int{2024, 2023, 2022, 2021, 2020},
	},
	SC: {
		ID:            SC,
		Name:          "情報処理安全確保支援士",
		ShortName:     "SC",
		OfficialTitle: "情報処理安全確保支援士試験（登録セキスペ）",
		PassingScore:  60,
		CategoryWeights: map[string]float64{
			"暗号・認証技術":       0.25,
			"攻撃手法と脆弱性対策":    0.30,
			"ネットワーク・OSセキュリティ": 0.20,
			"セキュアプログラミング":   0.15,
			"セキュリティマネジメント・法規": 0.10,
		},
		SyllabusTopics: []string{
			"暗号技術（公開鍵暗号, 共通鍵暗号, 電子署名, PKI）",
			"認証（OAuth 2.0, OpenID Connect, SAML, FIDO2）",
			"攻撃手法（SQLi, XSS, CSRF, ランサムウェア）",
			"侵入検知・防御（WAF, IDS/IPS, EDR, SIEM）",
			"セキュア開発ライフサイクル（SSDL）",
		},
		IPACode:        "sc",
		SupportedYears: []int{2024, 2023, 2022, 2021, 2020},
	},
}

// DefaultExamID specifies the default exam target if none is specified in CLI.
const DefaultExamID = DB

// Lookup finds an exam definition by case-insensitive key.
func Lookup(key string) (Definition, bool) {
	k := ID(strings.ToLower(strings.TrimSpace(key)))
	def, exists := Registry[k]
	return def, exists
}

// AvailableExams returns a slice of all registered exam definitions.
func AvailableExams() []Definition {
	var list []Definition
	for _, def := range Registry {
		list = append(list, def)
	}
	return list
}

// ValidateExamKey checks if a string is a valid exam key or returns default.
func ValidateExamKey(key string) (ID, error) {
	if strings.TrimSpace(key) == "" {
		return DefaultExamID, nil
	}
	k := ID(strings.ToLower(strings.TrimSpace(key)))
	if _, ok := Registry[k]; ok {
		return k, nil
	}
	return "", fmt.Errorf("unknown exam '%s' (available: db, nw, sc, pm, sa)", key)
}

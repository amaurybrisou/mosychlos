package models

import (
	"encoding/json"
	"testing"
)

func TestInvestmentResearchUnmarshaling(t *testing.T) {
	// Test JSON from actual AI response that was causing unmarshaling errors
	jsonData := `{
  "executive_summary": {
    "time_horizon": "3-7 ans (croissance)",
    "market_outlook": "Modérément favorable aux actions de croissance, avec volatilité persistante sur les marchés financiers et forte dispersion entre secteurs technologiques et valeurs cycliques ; les cryptomonnaies restent très volatiles et corrélées aux flux de risque global (court terme). (sources: yfinance_market_data, news_api)",
    "key_takeaways": [
      "Votre portefeuille est actuellement sur-concentré en crypto (100%): exposition élevée au risque idiosyncratique et volatilité extrême.",
      "Pour une stratégie growth européenne en EUR, privilégier des ETF actions européennes PEA-compatibles et small/mid caps pour le potentiel de surperformance.",
      "Maintenir une allocation crypto limitée (5–15%) et l’implémenter progressivement via DCA ; sécuriser liquidités et diversification sectorielle.",
      "Surveiller les décisions de la BCE et l’évolution de l’inflation en zone euro — elles dictent le profil rendement/risque des actions et des obligations (sources: news_api, fred)."
    ],
    "recommended_actions": [
      "Réduire progressivement la part crypto de ~100% à ~10% du portefeuille cible via ventes par tranches (DCA out).",
      "Allouer ~60–75% à actions growth EUR/global (ETFs PEA-compatibles + sélection de small/mid caps), 10–20% obligations courtes ou cash EUR, 5–10% crypto (exposition BNB conservée si conviction), le reste liquidités.",
      "Utiliser PEA pour positions actions européennes éligibles et assurance-vie pour diversifier en USD/ETFs non-PEA si nécessaire (optimisation fiscale).",
      "Mettre en place règles de gestion: rééquilibrage trimestriel, seuil de concentration max 10–15% par actif, stop-loss adaptatif pour crypto (ex.: -30% par tranche) et prise de profits par paliers."
    ]
  },
  "market_analysis": {
    "economic_backdrop": "Croissance moyenne en Europe avec variabilité sectorielle ; la trajectoire des taux de la BCE reste le facteur clef pour la rotation entre value et growth. (source: news_api)",
    "currency_impact": "Forte préférence pour instruments EUR afin de réduire risque de change ; si exposition USD/tech US, prévoir couverture ou placement en wrappers fiscalement adaptés (assurance-vie, compte-titres).",
    "liquidity_conditions": "Liquidité élevée pour ETF large cap européens ; cryptos très liquides mais avec slippage et spreads plus larges en période volatile. (source: yfinance_market_data)",
    "market_volatility": 0.35,
    "overall_sentiment": "Prudente à constructive pour les actions de croissance, risk-on pour certains segments technologiques ; crypto sentiment reste spéculatif. (sources: yfinance_market_data, news_api)",
    "sector_performance": {
      "Technologie": "Surperformance relative (momentum) ces derniers mois — attractif pour croissance mais valorisations élevées. (source: yfinance_market_data)",
      "Energie": "Volatile, joue le rôle de diversificateur selon cycle.",
      "Finances": "Mixed; sensibles aux taux et aux spreads."
    },
    "valuation_levels": "Valuations élevées sur les leaders tech US/quelques valeurs growth européennes ; meilleure opportunité dans small/mid caps européennes décotées par rapport au potentiel de croissance.",
    "technical_indicators": {
      "indices_trend": "Indices européens modérément haussiers ces 3 mois, Nasdaq en tête (source: yfinance_market_data).",
      "crypto_trend": "Volatilité élevée, mouvements par vagues spéculatives (source: yfinance_market_data)."
    }
  },
  "investment_themes": [
    {
      "name": "Croissance technologique européenne & digitalisation",
      "description": "Exposition aux leaders et aux challengers européens de la digitalisation, cybersécurité et semi-conducteurs locaux.",
      "key_drivers": [
        "Transformation numérique post-COVID",
        "Investissements privés et publics en sécurité digitale"
      ],
      "regional_exposure": {
        "country": "UE",
        "notes": "Favoriser titres éligibles PEA si possible"
      },
      "time_horizon": "3-7 ans",
      "recommended_allocation": "20–35% du portefeuille actions",
      "access_methods": [
        "ETF Growth Europe (PEA éligible si disponible)",
        "Sélection de small/mid caps via PEA ou CTO"
      ],
      "regulatory_support": true,
      "growth_projection": "Supérieur au marché européen de base sur 3-5 ans, mais dépend des cycles de valorisation"
    },
    {
      "name": "Small & Mid Caps européennes (croissance)",
      "description": "Capitaliser sur potentiel d’innovation et de rattrapage de valorisation.",
      "key_drivers": [
        "Reprise économique, niches industrielles",
        "Moindre couverture analytique -> prime de performance possible"
      ],
      "time_horizon": "4-8 ans",
      "recommended_allocation": "15–25% du portefeuille actions",
      "access_methods": [
        "ETF small/mid caps Europe (vérifier PEA)",
        "Fonds actifs spécialisés"
      ],
      "regulatory_support": false,
      "growth_projection": "Plus volatile mais avec potentiel supérieur"
    },
    {
      "name": "Allocation prudente en crypto (tail allocation)",
      "description": "Maintenir une exposition ciblée en cryptos à titre de 'satellite' growth, limitée par rapport à l’allocation actuelle.",
      "key_drivers": [
        "Adoption, régulation (MiCA en Europe), innovations NFT/DeFi",
        "Flux sur stablecoins et tokenomics"
      ],
      "time_horizon": "1-5 ans (très volatile)",
      "recommended_allocation": "5–10% portefeuille total (max 15% si forte tolérance)",
      "access_methods": [
        "Exposition spot via courtier régulé",
        "Produits listés ou ETF crypto si disponibles en zone euro"
      ],
      "regulatory_support": false,
      "growth_projection": "Haut potentiel, haut risque"
    }
  ],
  "research_findings": [
    {
      "title": "Diversifier hors crypto vers actions européennes (ETF PEA) et small/mid caps",
      "asset_class": "ETF Actions / Small & Mid Caps",
      "geographic_focus": "Europe (EUR)",
      "investment_theme": "Croissance, digitalisation, reprise domestique",
      "expected_return": {
        "base_case": 0.08,
        "confidence": "moyenne",
        "methodology": "Rebalancing vers ETF broad + sélection small/mid augmente espérance de rendement tout en réduisant volatilité relative",
        "time_horizon": "3-7 ans"
      },
      "valuation_metrics": {},
      "market_drivers": [
        "Politiques BCE",
        "Croissance des bénéfices européennes"
      ],
      "risk_profile": {
        "concentration_risk": "réduit si diversification ETF",
        "currency_risk": "faible pour EUR-denominated",
        "liquidity_risk": "faible pour large-cap ETF, moyen pour small/mid",
        "volatility_estimate": 0.18
      },
      "local_availability": true,
      "specific_instruments": [
        {
          "name": "ETF large-cap Européen (UCITS, EUR)",
          "type": "ETF",
          "currency": "EUR",
          "accessibility_notes": [
            "Vérifier PEA-compatibilité du produit avant allocation"
          ],
          "ira_eligible": false,
          "pea_eligible": true,
          "tfsa_eligible": false,
          "pea_eligible_note": "PEA éligible si ETF investit majoritairement en actions EU/EEE"
        }
      ],
      "tax_implications": [
        "Actions détenues via PEA: exonération d’impôt sur gains après 5 ans (contributions sociales restent dues à la sortie en capital selon règles).",
        "CTO/Assurance-vie: mécanismes différents, prévoir optimisation selon horizon."
      ],
      "time_horizon": "3-7 ans"
    },
    {
      "title": "Réduction structurée de la position crypto actuelle",
      "asset_class": "Crypto (PEPE, BNB)",
      "geographic_focus": "Global (marchés crypto)",
      "investment_theme": "Allocation satellite pour alpha spéculatif",
      "expected_return": {
        "base_case": 0.25,
        "confidence": "faible (forte dispersion)",
        "methodology": "Réduction progressive et détention d’une petite allocation conservatrice en crypto pour upside",
        "time_horizon": "1-5 ans"
      },
      "valuation_metrics": {},
      "market_drivers": [
        "Adoption, régulation (MiCA)",
        "Flux vers grands tokens (BNB vs tokens memecoin)"
      ],
      "risk_profile": {
        "concentration_risk": "élevée si non réduite",
        "currency_risk": "exposition USD/crypto native",
        "liquidity_risk": "variable selon token",
        "volatility_estimate": 0.8
      },
      "local_availability": true,
      "specific_instruments": [
        {
          "name": "BNB (Spot via courtier régulé)",
          "type": "Crypto",
          "currency": "USD/crypto",
          "accessibility_notes": [
            "Préférer courtiers régulés en France/UE; vérifier fiscalité crypto"
          ],
          "ira_eligible": false,
          "pea_eligible": false,
          "tfsa_eligible": false
        }
      ],
      "tax_implications": [
        "En France, les plus-values sur crypto pour particuliers sont soumises au régime spécifique (PFU/flat tax sur les plus-values sur actifs numériques pour les personnes non-professionnelles) — vérifier cas particulier et conservations (source: actualités réglementaires via news_api)."
      ],
      "time_horizon": "1-5 ans"
    }
  ],
  "risk_considerations": [
    {
      "type": "Macro (Taux et inflation)",
      "probability": "moyenne",
      "impact": "élevé",
      "severity": "peut induire rotation sectorielle et baisse des valeurs growth",
      "timeline": "0-12 mois",
      "mitigation": "diversifier sectoriellement, inclure un coussin d’obligations courtes ou cash"
    },
    {
      "type": "Concentration (crypto)",
      "probability": "élevée",
      "impact": "très élevé (risque de perte permanente de capital)",
      "severity": "très élevé",
      "timeline": "immédiat",
      "mitigation": "réduire la position progressivement, règles de stop-loss et prise de profits par paliers"
    },
    {
      "type": "Réglementaire (MiCA & taxes franco-européennes)",
      "probability": "moyenne",
      "impact": "moyen à élevé pour crypto et produits non-UE",
      "severity": "moyen",
      "timeline": "0-24 mois",
      "mitigation": "préférer plateformes régulées UE, rester informé via sources officielles (news_api)"
    }
  ],
  "implementation_plan": {
    "step_by_step": [
      "1) Évaluer comptes disponibles: ouvrir PEA si pas encore (pour exonération après 5 ans) ; garder CTO/assurance-vie pour ETF non-PEA et exposition US.",
      "2) Vendre progressivement PEPE (ex.: 25% de la position toutes les 2 semaines) jusqu’à atteindre cible crypto 5–10% du portefeuille. Convertir part en EUR cash/ETF pendant rebalancing.",
      "3) Allouer les fonds dégagés vers: 60–75% actions (via ETF Europe PEA + small/mid caps), 10–20% obligations courtes ou cash court terme (fonds monétaire court).",
      "4) Mise en place de DCA mensuel pour nouvelles positions (pour lisser le prix d’entrée), taille des ordres adaptée au capital actuel (ex.: 50 EUR/mois jusqu’à 3–6 mois).",
      "5) Rebalancer quarterly; surveiller seuils: toute position >15% déclenche prise de bénéfice partielle.",
      "6) Suivi & reporting: revoir performance tous les 3 mois et ajuster en fonction des décisions BCE et news macro (source: news_api)."
    ],
    "position_size": "Étant donné la taille du portefeuille (EUR 432), privilégier micro-ordres et DCA ; ne pas engager >10–15% de capital total sur un trade isolé.",
    "entry_strategy": "DCA pour nouvelles allocations ; tranches fixes pour réduction crypto.",
    "exit_criteria": [
      "Crypto: revente supplémentaire si drawdown >50% après rééquilibrage ou si token perd signaux fondamentaux/techniques.",
      "Actions growth: prendre profits partiels si valuation relative dépasse historique de secteur ou si indicateurs macro se dégradent fortement.",
      "Stop-loss: appliquer règles strictes pour positions spéculatives (ex.: 20–30% en dessous du prix d’achat pour petites positions, adaptée selon volatilité)."
    ],
    "monitoring_points": [
      "Décisions de politique monétaire de la BCE et communiqués (source: news_api).",
      "Données macro européennes et inflation (source: news_api, fred pour séries historiques).",
      "Évolution des prix et volumes des tokens détenus (source: yfinance_market_data).",
      "Performance relative ETFs Europe vs US (source: yfinance_market_data)."
    ],
    "timeline": "Déploiement 3-6 mois pour atteindre allocation cible ; revues trimestrielles ensuite."
  },
  "regional_considerations": {
    "country": "FR",
    "currency_focus": "EUR",
    "language": "fr",
    "local_market_access": [
      {
        "exchange": "Euronext Paris",
        "asset_classes": [
          "Actions (large, mid, small)",
          "ETF UCITS",
          "Obligations d’entreprise listées"
        ],
        "trading_hours": "09:00–17:30 CET",
        "trading_costs": "Frais de courtage locaux et spreads ; privilégier courtiers à faible coût pour petits montants"
      }
    ],
    "tax_optimizations": [
      {
        "account_type": "PEA",
        "strategy": "Loger actions européennes/ETF PEA-compatibles pour exonération d’impôt sur les gains après 5 ans (optimiser pour horizon long).",
        "implementation": "Transférer positions actions éligibles sur PEA si possible et fiscalement avantageux; sinon utiliser CTO/assurance-vie."
      },
      {
        "account_type": "Assurance-vie",
        "strategy": "Utiliser pour ETF non-PEA ou exposition USD ; avantage fiscaux à long terme et transmission.",
        "implementation": "Allouer parts défensives ou internationales via wrappers assurance-vie selon plafond et frais."
      }
    ]
  },
  "actionable_insights": [
    {
      "instrument": {
        "type": "ETF (Europe, UCITS, EUR)",
        "name": "ETF large-cap Europe (PEA-eligible)",
        "currency": "EUR",
        "pea_eligible": true,
        "isa_eligible": false,
        "tfsa_eligible": false,
        "accessibility_notes": [
          "Choisir ETF UCITS, réplication physique, faible TER (<0.3%)",
          "Vérifier PEA-compatibilité avant l’achat"
        ]
      },
      "priority": "Haute",
      "position_size": "60–75% des avoirs actions (soit ~35–55% du portefeuille total cible)",
      "target_allocation": "Allouer en priorité via PEA",
      "entry_strategy": "DCA mensuel (ex.: 50 EUR/mois si possible) ou achats échelonnés si conversion de crypto en cash",
      "exit_criteria": [
        "Prendre profits partiels si l’ETF surperforme >30% en 12 mois et si rotation macro défavorable",
        "Rééquilibrage trimestriel"
      ],
      "monitoring_points": [
        "Frais TER, liquidité de l’ETF, éligibilité PEA",
        "Performance relative vs STOXX50E (source: yfinance_market_data)"
      ],
      "rationale": "Permet une exposition growth européenne tout en respectant contrainte EUR/PEA et en réduisant concentration crypto"
    },
    {
      "instrument": {
        "type": "Small/Mid Cap Fund or ETF (Europe)",
        "name": "ETF small/mid caps Europe (vérifier PEA)",
        "currency": "EUR",
        "pea_eligible": true,
        "accessibility_notes": [
          "Favoriser fonds actifs/ETF avec historique de performance et liquidité suffisante"
        ]
      },
      "priority": "Moyenne",
      "position_size": "15–25% des avoirs actions",
      "target_allocation": "Augmente la probabilité d’alpha sur 3-7 ans",
      "entry_strategy": "Paliers et DCA (réduire le risque d’achat au sommet)",
      "exit_criteria": [
        "Réduire si performance en conservation stagne sur 12–18 mois ou si détérioration fondamentale"
      ],
      "monitoring_points": [
        "Couverture analytique, news sectorielles (source: news_api)"
      ],
      "rationale": "Potentiel de rendement supérieur aux large caps avec une prime de risque acceptable pour profil growth"
    },
    {
      "instrument": {
        "type": "Crypto (spot)",
        "name": "BNB (conserver partiellement) + réduire PEPE",
        "currency": "crypto",
        "pea_eligible": false,
        "accessibility_notes": [
          "Utiliser courtier régulé UE/FR",
          "Conserver journaux de transactions pour déclaration fiscale"
        ]
      },
      "priority": "Moyenne/Spéculative",
      "position_size": "5–10% du portefeuille total (après rééquilibrage)",
      "target_allocation": "Réduire PEPE fortement; garder BNB uniquement si conviction long-term",
      "entry_strategy": "Réaliser ventes par tranches pour réduire exposition",
      "exit_criteria": [
        "Vendre si token perd support technique majeur ou si régulation défavorable",
        "Reprendre position uniquement sur convinction fondamentale renforcée"
      ],
      "monitoring_points": [
        "Évolution réglementaire MiCA, annonces Binance/BNB, volumes et liquidité (source: news_api, yfinance_market_data)"
      ],
      "rationale": "Réduire risque de concentration tout en conservant une exposure spéculative calibrée"
    },
    {
      "instrument": {
        "type": "Liquidités / Obligations court terme",
        "name": "Fonds monétaires court terme / CETE",
        "currency": "EUR",
        "pea_eligible": false,
        "accessibility_notes": [
          "Utilisable via assurance-vie ou CTO pour lisser la volatilité"
        ]
      },
      "priority": "Moyenne",
      "position_size": "10–20% du portefeuille",
      "target_allocation": "Coussin de sécurité et opportunités marché",
      "entry_strategy": "Maintenir liquidités disponibles issues des ventes crypto",
      "exit_criteria": ["Réallouer vers actions sur corrections >8–10%"],
      "monitoring_points": [
        "Taux court terme en zone euro, décisions BCE (source: news_api, fred)"
      ],
      "rationale": "Réduit risque global et offre réserve d’achat lors de corrections"
    }
  ],
  "sources": [
    {
      "source": "yfinance_market_data",
      "title": "Indices & secteur performance (3 mois)",
      "url": "",
      "search_query": "yfinance market indices 3mo ^GSPC ^STOXX50E BTC-USD",
      "relevance_score": 0.9
    },
    {
      "source": "news_api",
      "title": "Actualités BCE, régulation crypto et macro Europe",
      "url": "",
      "search_query": "ECB Europe economy cryptocurrency regulation BNB PEPE",
      "relevance_score": 0.9
    },
    {
      "source": "fred",
      "title": "Series GeoFRED (Per Capita Personal Income - context historique)",
      "url": "",
      "search_query": "FRED GeoFRED series 882 per capita personal income",
      "relevance_score": 0.6
    }
  ],
  "metadata": {
    "generated_at": "2025-08-17T23:13:05+02:00",
    "engine_version": "v2.0",
    "research_depth": "advanced",
    "region": "FR"
  }
}
`

	var result InvestmentResearchResult
	err := json.Unmarshal([]byte(jsonData), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// Verify the data was parsed correctly
	if len(result.InvestmentThemes) == 0 {
		t.Fatal("Expected at least one investment theme")
	}

	theme := result.InvestmentThemes[0]
	if theme.Name != "Croissance technologique européenne & digitalisation" {
		t.Errorf("Expected theme name 'Croissance technologique européenne & digitalisation', got '%s'", theme.Name)
	}

	if len(theme.RegionalExposure) == 0 {
		t.Fatal("Expected regional exposure data")
	}

	t.Logf("Successfully parsed investment theme: %+v", theme)
}

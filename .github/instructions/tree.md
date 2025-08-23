.
├── build
│   └── mosychlos
├── cmd
│   └── mosychlos
│   ├── analyze.go
│   ├── batch.go
│   ├── main.go
│   ├── portfolio.go
│   └── tools.go
├── config
│   ├── config.default.yaml
│   ├── config.yaml
│   ├── investment_profiles
│   │   └── defaults
│   │   ├── CA
│   │   ├── FR
│   │   │   ├── aggressive.yaml
│   │   │   ├── conservative.yaml
│   │   │   └── moderate.yaml
│   │   ├── global
│   │   │   └── moderate.yaml
│   │   └── US
│   │   ├── aggressive.yaml
│   │   └── moderate.yaml
│   └── templates
│   └── investment_research
│   ├── base
│   │   ├── components
│   │   │   ├── context.tmpl
│   │   │   ├── output_format.tmpl
│   │   │   ├── portfolio_analysis.tmpl
│   │   │   └── research_framework.tmpl
│   │   └── research.tmpl
│   └── regional
│   ├── localization
│   │   ├── AU_en.yaml
│   │   ├── CA_en.yaml
│   │   ├── CH_en.yaml
│   │   ├── DE_de.yaml
│   │   ├── DE_en.yaml
│   │   ├── FR_fr.yaml
│   │   ├── GLOBAL_en.yaml
│   │   ├── JP_en.yaml
│   │   ├── NL_en.yaml
│   │   ├── SG_en.yaml
│   │   ├── UK_en.yaml
│   │   └── US_en.yaml
│   └── overlays
│   ├── CA_overlay.tmpl
│   ├── FR_overlay.tmpl
│   └── US_overlay.tmpl
├── debug
│   └── debug_schema.go
├── docker
│   └── finrobot
│   └── Dockerfile
├── docker-compose.yaml
├── Dockerfile
├── docs
│   ├── ai-architecture-refactoring-proposal.md
│   ├── ai-batch-processing-architecture.md
│   ├── batch_llm_function_implementation_guide.md
│   ├── engine-chaining-single-context-analysis.md
│   ├── finrobot-architecture-reference.md
│   ├── finrobot-data-sources.md
│   ├── finrobot-enhancement-plan.md
│   ├── finrobot-personas-implementation.md
│   ├── finrobot-prompts-implementation.md
│   ├── finrobot-tools-implementation.md
│   ├── gpt-5.md
│   ├── investment-profile-manager.md
│   ├── investment-research-engine.md
│   ├── investment-research-technical-corrections.md
│   ├── llm_batch_processing_architecture-1.md
│   ├── llm_batch_processing_architecture.md
│   ├── llm-batch-test-implementation.md
│   ├── openai-rate-limiting-architecture.md
│   ├── openai-websearch.md
│   ├── portfolio_prompts_by_region.md
│   ├── prompt-system-refactoring.md
│   ├── README.md
│   ├── regional-investment-opportunities-implementation.md
│   ├── single-context-multi-agent-analysis.md
│   ├── tool-driven-professional-roles.md
│   └── tools-inventory.md
├── functions.txt
├── go.mod
├── go.sum
├── internal
│   ├── adapters
│   │   └── binance.go
│   ├── ai
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── factory.go
│   │   ├── openai
│   │   │   ├── provider.go
│   │   │   ├── provider_test.go
│   │   │   ├── response-api.go
│   │   │   ├── response_api_test.go
│   │   │   ├── session.go
│   │   │   ├── stream.go
│   │   │   ├── stream_test.go
│   │   │   └── web_search_preview_tracking.go
│   │   ├── README.md
│   │   └── schema.go
│   ├── budget
│   │   ├── consumer.go
│   │   ├── consumer_test.go
│   │   └── README.md
│   ├── config
│   │   ├── config.go
│   │   ├── config_test.go
│   │   └── types.go
│   ├── engine
│   │   ├── batch_investment_research
│   │   │   └── engine.go
│   │   ├── compliance
│   │   │   └── compliance.go
│   │   ├── investment_research
│   │   │   └── engine.go
│   │   ├── orchestrator.go
│   │   ├── README.md
│   │   └── risk
│   │   └── risk.go
│   ├── health
│   │   ├── monitor.go
│   │   └── README.md
│   ├── jurisdiction
│   │   ├── apply.go
│   │   ├── policy.json
│   │   ├── README.md
│   │   ├── service.go
│   │   └── service_test.go
│   ├── llm
│   │   ├── batch
│   │   │   ├── cli_integration.go
│   │   │   ├── cost_optimizer.go
│   │   │   ├── cost_optimizer_test.go
│   │   │   ├── manager.go
│   │   │   ├── monitor.go
│   │   │   ├── result_aggregator.go
│   │   │   └── result_aggregator_test.go
│   │   ├── client.go
│   │   ├── factory.go
│   │   ├── openai
│   │   │   ├── batch_client.go
│   │   │   ├── batch_formatter.go
│   │   │   └── batch_formatter_test.go
│   │   ├── README.md
│   │   └── util.go
│   ├── localization
│   │   └── service.go
│   ├── portfolio
│   │   ├── interface.go
│   │   ├── portfolio_test.go
│   │   ├── README.md
│   │   ├── service.go
│   │   └── validator.go
│   ├── profile
│   │   ├── manager.go
│   │   ├── manager_test.go
│   │   └── README.md
│   ├── prompt
│   │   ├── interface.go
│   │   ├── loader.go
│   │   ├── manager.go
│   │   ├── regional_manager.go
│   │   ├── regional_overlay.go
│   │   ├── templates
│   │   │   ├── investment_research
│   │   │   │   ├── base
│   │   │   │   │   └── components
│   │   │   │   └── regional
│   │   │   │   ├── localization
│   │   │   │   └── overlays
│   │   │   ├── market
│   │   │   │   └── outlook.tmpl
│   │   │   ├── portfolio
│   │   │   │   ├── allocation.tmpl
│   │   │   │   ├── compliance.tmpl
│   │   │   │   ├── performance.tmpl
│   │   │   │   ├── reallocation.tmpl
│   │   │   │   └── risk.tmpl
│   │   │   └── README.md
│   │   └── types.go
│   ├── report
│   │   ├── data.go
│   │   ├── fixtures
│   │   │   └── investment_research.sample.json
│   │   ├── generator.go
│   │   ├── README.md
│   │   ├── renderer.go
│   │   ├── templates
│   │   │   ├── customer.md
│   │   │   ├── investment_research.md
│   │   │   ├── system.md
│   │   │   └── web_search_preview.md
│   │   └── types.go
│   └── tools
│   ├── cached_tool.go
│   ├── cached_tool_test.go
│   ├── fmp
│   │   ├── fmp.go
│   │   ├── tool.go
│   │   └── tools_config.go
│   ├── fmp_estimates
│   │   ├── analyst_estimates.go
│   │   ├── analyst_estimates_test.go
│   │   └── tools_config.go
│   ├── fmp_tool_test.go
│   ├── fred
│   │   ├── provider.go
│   │   ├── README.md
│   │   ├── tools_config.go
│   │   └── tool_test.go
│   ├── fred_tool_test.go
│   ├── metrics_wrapper.go
│   ├── modular_config_test.go
│   ├── newsapi
│   │   ├── newsapi.go
│   │   ├── newsapi_test.go
│   │   └── tools_config.go
│   ├── newsapi_tool_test.go
│   ├── rate_limiter.go
│   ├── README.md
│   ├── sec_edgar
│   │   ├── sec_edgar.go
│   │   ├── sec_edgar_test.go
│   │   └── tools_config.go
│   ├── tool_registry.go
│   ├── tools.go
│   ├── weather
│   │   └── weather.go
│   └── yfinance
│   ├── dividends.go
│   ├── financials.go
│   ├── market_data.go
│   ├── README.md
│   ├── stock_data.go
│   ├── stock_info.go
│   ├── tools_config.go
│   └── yfinance_test.go
├── main.go
├── mosychlos-cache
│   └── tools
│   ├── 2025-08-16
│   │   ├── tool_fmp_2025-08-16_1fc5f3516912eb8c.json
│   │   ├── tool_fmp_2025-08-16_9d1604e267d7585d.json
│   │   ├── tool_fmp_2025-08-16_b23f5e72175413e4.json
│   │   ├── tool_fmp_analyst_estimates_2025-08-16_01d93dc47145ae0f.json
│   │   ├── tool_fmp_analyst_estimates_2025-08-16_ec6629b9b9bfaa67.json
│   │   ├── tool_news_api_2025-08-16_043566bfce3f359f.json
│   │   ├── tool_news_api_2025-08-16_0cf5e149146688d3.json
│   │   ├── tool_news_api_2025-08-16_0fc51992a31e5783.json
│   │   ├── tool_news_api_2025-08-16_1c047e8c943d9236.json
│   │   ├── tool_news_api_2025-08-16_30b714a25caff2c7.json
│   │   ├── tool_news_api_2025-08-16_3510a820e103b84b.json
│   │   ├── tool_news_api_2025-08-16_5f6ba0bc5450cc34.json
│   │   ├── tool_news_api_2025-08-16_7876e692a18ff6cd.json
│   │   ├── tool_news_api_2025-08-16_7c4e2267407c98fa.json
│   │   ├── tool_news_api_2025-08-16_9ec7b5e999f56931.json
│   │   ├── tool_news_api_2025-08-16_c510b75313415cfd.json
│   │   ├── tool_news_api_2025-08-16_c7541212f7a9a045.json
│   │   ├── tool_news_api_2025-08-16_d534de2d8ef0b797.json
│   │   ├── tool_news_api_2025-08-16_d7c64ef26c3f2ea5.json
│   │   ├── tool_news_api_2025-08-16_dad80637b467d3af.json
│   │   ├── tool_news_api_2025-08-16_de33aae2802f335c.json
│   │   ├── tool_news_api_2025-08-16_e361243cc47c3cde.json
│   │   ├── tool_news_api_2025-08-16_f09ba0a6bb87b076.json
│   │   ├── tool_news_api_2025-08-16_f547c8cd515d207d.json
│   │   ├── tool_yfinance_market_data_2025-08-16_48aad3f008eaa988.json
│   │   ├── tool_yfinance_market_data_2025-08-16_4b463f152de927ea.json
│   │   ├── tool_yfinance_stock_data_2025-08-16_0f858309396f1746.json
│   │   ├── tool_yfinance_stock_data_2025-08-16_103887150dfe6513.json
│   │   ├── tool_yfinance_stock_data_2025-08-16_15bc723ceac2175e.json
│   │   ├── tool_yfinance_stock_data_2025-08-16_2d684ac8096875ae.json
│   │   ├── tool_yfinance_stock_data_2025-08-16_4ee5b085687223a6.json
│   │   ├── tool_yfinance_stock_data_2025-08-16_53c49312499b7016.json
│   │   ├── tool_yfinance_stock_data_2025-08-16_64d638d8744bce9e.json
│   │   ├── tool_yfinance_stock_data_2025-08-16_9ae8c0a3f425b310.json
│   │   ├── tool_yfinance_stock_data_2025-08-16_ac5d9a4da52c60a6.json
│   │   ├── tool_yfinance_stock_data_2025-08-16_c8fc8ca500b535c3.json
│   │   ├── tool_yfinance_stock_info_2025-08-16_81f2f55d6b43e79f.json
│   │   └── tool_yfinance_stock_info_2025-08-16_e49392660bcfb7da.json
│   └── 2025-08-17
│   ├── tool_fmp_2025-08-17_4388d0882e828eb4.json
│   ├── tool_fmp_2025-08-17_6a827f892bf2fd70.json
│   ├── tool_fmp_2025-08-17_df101b94e8e08be9.json
│   ├── tool_fmp_analyst_estimates_2025-08-17_df101b94e8e08be9.json
│   ├── tool_fred_2025-08-17_354b1ca6b807a08c.json
│   ├── tool_news_api_2025-08-17_04093ce64debe068.json
│   ├── tool_news_api_2025-08-17_1cc91db11477cb13.json
│   ├── tool_news_api_2025-08-17_41569e20627480ab.json
│   ├── tool_news_api_2025-08-17_5099321fe3a312a0.json
│   ├── tool_news_api_2025-08-17_5bd5a7a33906bbce.json
│   ├── tool_news_api_2025-08-17_67f93fb3b5445c02.json
│   ├── tool_news_api_2025-08-17_86f44b99a2a96bcb.json
│   ├── tool_news_api_2025-08-17_9c7a2d7852fcbe54.json
│   ├── tool_news_api_2025-08-17_ad24eb367063709f.json
│   ├── tool_news_api_2025-08-17_b90220caed7add59.json
│   ├── tool_news_api_2025-08-17_c6393919d21bc654.json
│   ├── tool_news_api_2025-08-17_d6279f3a354110ac.json
│   ├── tool_news_api_2025-08-17_d8ca63adaa4369e8.json
│   ├── tool_yfinance_market_data_2025-08-17_31a821e3f2ec5d78.json
│   ├── tool_yfinance_market_data_2025-08-17_bb9a87a50eb03306.json
│   ├── tool_yfinance_market_data_2025-08-17_ebf9943e50bc595f.json
│   ├── tool_yfinance_market_data_2025-08-17_f2ed65c02994b503.json
│   └── tool_yfinance_stock_data_2025-08-17_d3f48d380d9e97a3.json
├── mosychlos-data
│   ├── batch
│   │   ├── job_batch_68a22e45531c81908bad1f9d0b6a7b26.json
│   │   ├── job_batch_68a2356e7124819089128a080763a710.json
│   │   ├── results_batch_68a22d8e2fe881908d345575f8e03f94.json
│   │   └── results_batch_68a2356e7124819089128a080763a710.json
│   ├── portfolio
│   │   ├── current.yaml
│   │   └── empty_test.yaml
│   └── reports
│   ├── full_report_20250817_203748.md
│   ├── full_report_20250817_203750.json
│   ├── full_report_20250817_203750.pdf
│   ├── full_report_20250817_232242.md
│   ├── full_report_20250817_232243.json
│   └── full_report_20250817_232243.pdf
├── pkg
│   ├── bag
│   │   ├── bag.go
│   │   ├── bag_test.go
│   │   ├── README.md
│   │   └── shared_bag.go
│   ├── binance
│   │   ├── binance_test.go
│   │   ├── client.go
│   │   ├── interface.go
│   │   ├── portfolio.go
│   │   └── README.md
│   ├── cache
│   │   ├── cache.go
│   │   ├── cache_test.go
│   │   ├── example_test.go
│   │   ├── file_cache_bench_test.go
│   │   ├── file_cache.go
│   │   ├── file_cache_test.go
│   │   ├── monitor.go
│   │   └── README.md
│   ├── cli
│   │   ├── analyze.go
│   │   ├── display.go
│   │   ├── portfolio.go
│   │   ├── prompt.go
│   │   └── report.go
│   ├── config
│   │   └── loader.go
│   ├── errors
│   │   ├── context.go
│   │   ├── data.go
│   │   ├── llm.go
│   │   ├── profile.go
│   │   ├── prompt.go
│   │   ├── README.md
│   │   └── rebalance.go
│   ├── fmp
│   │   └── client.go
│   ├── fred
│   │   └── client.go
│   ├── fs
│   │   ├── fs.go
│   │   └── README.md
│   ├── keys
│   │   ├── bag.go
│   │   └── README.md
│   ├── log
│   │   ├── log.go
│   │   ├── log_test.go
│   │   └── README.md
│   ├── models
│   │   ├── ai_batch.go
│   │   ├── ai.go
│   │   ├── analysis.go
│   │   ├── analyst_estimates.go
│   │   ├── binance.go
│   │   ├── economic.go
│   │   ├── economic_norm.go
│   │   ├── engine.go
│   │   ├── fmp.go
│   │   ├── fred.go
│   │   ├── fundamentals.go
│   │   ├── fundamentals_norm.go
│   │   ├── health_metrics.go
│   │   ├── investment_profile.go
│   │   ├── investment_profile_test.go
│   │   ├── investment_research.go
│   │   ├── investment_research_test.go
│   │   ├── jurisdiction.go
│   │   ├── jurisdiction_norm.go
│   │   ├── localization.go
│   │   ├── news.go
│   │   ├── news_norm.go
│   │   ├── normalization_test.go
│   │   ├── normalized.go
│   │   ├── portfolio.go
│   │   ├── portfolio_norm.go
│   │   ├── portfolio_test.go
│   │   ├── regional.go
│   │   ├── regional_test.go
│   │   ├── report.go
│   │   ├── sec.go
│   │   ├── tool_metrics.go
│   │   ├── validator.go
│   │   └── yfinance.go
│   ├── nativeutils
│   │   ├── ptr.go
│   │   └── ptr_test.go
│   ├── newsapi
│   │   ├── client.go
│   │   └── client_test.go
│   ├── pdf
│   │   └── pdf.go
│   ├── persist
│   │   ├── manager.go
│   │   ├── manager_test.go
│   │   └── README.md
│   ├── sec
│   │   └── client.go
│   └── yfinance
│   └── client.go
└── testdata
└── batch
├── errors.jsonl
├── output.jsonl
└── requests.jsonl

91 directories, 342 files

// Package toolsimpl
package toolsimpl

import (
	_ "github.com/amaurybrisou/mosychlos/internal/toolsimpl/fmp"
	_ "github.com/amaurybrisou/mosychlos/internal/toolsimpl/fred"
	_ "github.com/amaurybrisou/mosychlos/internal/toolsimpl/newsapi"
	_ "github.com/amaurybrisou/mosychlos/internal/toolsimpl/secedgar"
	_ "github.com/amaurybrisou/mosychlos/internal/toolsimpl/summarize"

	// _ "github.com/amaurybrisou/mosychlos/internal/toolsimpl/websearch"
	_ "github.com/amaurybrisou/mosychlos/internal/toolsimpl/yfinance"
)

/*
This module loads all the tools which individually Register themselves
into the `tools` package registry.
*/

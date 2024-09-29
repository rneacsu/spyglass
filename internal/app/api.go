package app

type AppApi struct {
	app *App
}

func (a *AppApi) GetGRPCUrl() string {
	return a.app.grpcServer.GetUrl()
}

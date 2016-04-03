package middle_test
import (
	"fmt"
	"github.com/orivil/middle"
)

//simulated actions
const (
	// AdminController's actions
	LoginAction  = "login"
	LogoutAction = "logout"
	RegisterAction = "register"

	// I18nController's actions
	SetLanguageAction = "setLanguage"
)

// simulated middleware
const (
	LoginMiddle = "login"
	LogoutMiddle = "logout"
	RegisterMiddle = "register"
	I18nMiddle = "I18n"
)

// controller name
const (
	AdminController = "adminController"
	I18nController = "i18nController"
)

// bundle name(package name)
const (
	AdminBundle = "admin"
	I18nBundle = "I18n"
)

// simulated controller
type controller struct {
	name string
	bundle string
	actions map[string]bool
}

var controllers = []controller{
	{
		bundle: AdminBundle, // package
		name: AdminController,
		actions: map[string]bool{
			LoginAction: true,
			LogoutAction: true,
			RegisterAction: true,
		},
	},

	{
		bundle: I18nBundle,
		name: I18nController,
		actions: map[string]bool{
			SetLanguageAction: true,
		},
	},
}

func ExampleBag() {
	// 1. new middleware bag
	bag := middle.NewMiddlewareBag()

	// 2. add all controllers
	for _, controller := range controllers {
		bag.AddController(controller.bundle, controller.name, controller.actions)
	}

	// 3. add middleware
	// set current package and controller first, this step should be auto set
	bag.SetCurrent(AdminBundle, AdminController)

	// set middleware
	bag.Set(LoginMiddle).AllBundles()
	bag.Set(LoginMiddle).ExceptActions(LogoutAction, RegisterAction) // except it


	bag.Set(LogoutMiddle).OnlyActions(LogoutAction)

	bag.Set(RegisterMiddle).OnlyActions(RegisterAction)

	// set current package and controller
	bag.SetCurrent(I18nBundle, I18nController)

	// set i18n middleware
	bag.Set(I18nMiddle).OnlyController(I18nController)

	// 4. get matched middleware
	action := AdminBundle+"."+AdminController+"."+LoginAction // get the full action name
	middles := bag.GetMiddles(action)
	fmt.Printf("action [%s] has middlewares: %v\n", action, middles)

	action = AdminBundle+"."+AdminController+"."+LogoutAction
	middles = bag.GetMiddles(action)
	fmt.Printf("action [%s] has middlewares: %v\n", action, middles)

	action = AdminBundle+"."+AdminController+"."+RegisterAction
	middles = bag.GetMiddles(action)
	fmt.Printf("action [%s] has middlewares: %v\n", action, middles)

	action = I18nBundle+"."+I18nController+"."+SetLanguageAction
	middles = bag.GetMiddles(action)
	fmt.Printf("action [%s] has middlewares: %v\n", action, middles)

	// Output:
	// action [admin.adminController.login] has middlewares: [login]
	// action [admin.adminController.logout] has middlewares: [logout]
	// action [admin.adminController.register] has middlewares: [register]
	// action [I18n.i18nController.setLanguage] has middlewares: [login I18n]
}
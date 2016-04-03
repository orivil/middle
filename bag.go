package middle
import (
	"log"
)

// the interface for middleware container, to check if middleware exist
type MiddleChecker interface {
	CheckExist(middleware string) error
}

type Bag struct {

	// {package: {controller: {action: true}}}
	actions map[string]map[string]map[string]bool

	// the matched middleware map
	// {package.controller.action: {middleware: true}}
	matchedMiddles map[string]map[string]bool

	checker MiddleChecker

	// current middleware
	_middleware string

	// current bundle
	_bundle string

	// current controller
	_controller string
}

func NewMiddlewareBag() *Bag {
	return &Bag{
		actions: make(map[string]map[string]map[string]bool, 100),
		matchedMiddles: make(map[string]map[string]bool, 10),
	}
}

func (b *Bag) addBundles(bundles []string) {
	for _, bundle := range bundles {
		controllers := b.actions[string(bundle)]
		for controller, actions := range controllers {
			b.addActions(bundle, controller, mapStrToSlice(actions))
		}
	}
}

func (b *Bag) deleteBundles(bundles []string) {
	for _, bundle := range bundles {
		controllers := b.actions[string(bundle)]
		for controller, actions := range controllers {
			b.deleteActions(bundle, controller, mapStrToSlice(actions))
		}
	}
}

func (b *Bag) addControllers(bundle string, controllers []string) {
	for _, controller := range controllers {
		actions := b.actions[string(bundle)][string(controller)]
		b.addActions(string(bundle), string(controller), mapStrToSlice(actions))
	}
}

func (b *Bag) deleteControllers(bundle string, controllers []string) {
	for _, controller := range controllers {
		actions := b.actions[string(bundle)][string(controller)]
		b.deleteActions(string(bundle), string(controller), mapStrToSlice(actions))
	}
}

func (b *Bag) addActions(bundle string, controller string, actions []string) {
	for _, action := range actions {
		b.addAction(bundle, controller, action)
	}
}

func (b *Bag) addAction(bundle string, controller string, action string) {
	action = bundle+"."+controller+"."+action
	if b.matchedMiddles[action] == nil {
		b.matchedMiddles[action] = map[string]bool{b._middleware: true}
	} else {
		b.matchedMiddles[action][b._middleware] = true
	}
}

func (b *Bag) deleteActions(bundle string, controller string, actions []string) {
	for _, action := range actions {
		delete(b.matchedMiddles[bundle+"."+controller+"."+action], b._middleware)
	}
}

func (b *Bag) checkBundles(bundles []string) {
	for _, bundle := range bundles {
		if _, ok := b.actions[bundle]; !ok {
			log.Fatalf("bundle %s: func SetMiddle got error, bundle %s not exist\n", bundle, bundle)
		}
	}
}

func (b *Bag) checkControllers(bundle string, controllers []string) {
	for _, controller := range controllers {
		if _, ok := b.actions[bundle][controller]; !ok {
			log.Fatalf("bundle %s: func SetMiddle got error, controller %s not exist\n", bundle, controller)
		}
	}
}

func (b *Bag) checkActions(bundle, controller string, actions []string) {
	for _, action := range actions {
		if !b.actions[bundle][controller][action] {
			if controller == "" {
				log.Fatalf("bundle %s: provider func SetMiddle got error, action middle must set in controller\n", bundle)
			} else {
				log.Fatalf("%s.%s: func SetMiddle got error, action %s not exist\n", bundle, controller, action)
			}
		}
	}
}

func (b *Bag) SetCurrent(bundle, controller string) {
	b._bundle = bundle
	b._controller = controller
}

// AddController add all controller messages
func (b *Bag) AddController(bundle string, controller string, actions map[string]bool) {
	if b.actions[bundle] == nil {
		b.actions[bundle] = map[string]map[string]bool{controller: actions}
	}
	b.actions[bundle][controller] = actions
}

// GetMiddles
//
// action: must be full name like "bundle.controller.action"
func (b *Bag) GetMiddles(action string) (matchedMiddles []string) {
	if ms, ok := b.matchedMiddles[action]; ok {
		matchedMiddles = make([]string, len(ms))
		index := 0
		for m, _ := range ms {
			matchedMiddles[index] = m
			index++
		}
	}
	return
}

// SetMiddleChecker for check if middleware exist
func (b *Bag) SetMiddleChecker(c MiddleChecker) {
	b.checker = c
}

func (b *Bag) Set(middleware string) *Bag {
	if b.checker != nil {
		err := b.checker.CheckExist(middleware)
		if err != nil {
			log.Fatal(err)
		}
	}
	b._middleware = middleware
	return b
}

func (b *Bag) AllBundles() *Bag {
	for bundle, controllers := range b.actions {
		for controller, actions := range controllers {
			b.addActions(bundle, controller, mapStrToSlice(actions))
		}
	}
	return b
}

func (b *Bag) AllControllers() {
	controllers := b.actions[b._bundle]
	for controller, actions := range controllers {
		b.addActions(b._bundle, controller, mapStrToSlice(actions))
	}
}

func (b *Bag) AllActions() {
	actions := b.actions[b._bundle][b._controller]
	b.addActions(b._bundle, b._controller, mapStrToSlice(actions))
}

func (b *Bag) OnlyBundle(bundles ...string) *Bag {
	b.checkBundles(bundles)

	// add matched bundle
	b.addBundles(bundles)

	// delete unmatched bundle
	reverseBundles := b.getReverseBundles(bundles)
	b.deleteBundles(reverseBundles)
	return b
}

func (b *Bag) ExceptBundle(bundles ...string) *Bag {
	b.checkBundles(bundles)

	// add matched bundle
	reverseBundles := b.getReverseBundles(bundles)
	b.addBundles(reverseBundles)

	// delete unmatched bundle
	b.deleteBundles(bundles)
	return b
}

func (b *Bag) OnlyController(controllers ...string) {
	b.checkControllers(b._bundle, controllers)

	// add matched controllers
	b.addControllers(b._bundle, controllers)

	// delete unmatched controllers
	reverseControllers := b.getReverseControllers(b._bundle, controllers)
	b.deleteControllers(b._bundle, reverseControllers)
}

func (b *Bag) ExceptController(controllers ...string) {
	b.checkControllers(b._bundle, controllers)

	// add matched controllers
	reverseControllers := b.getReverseControllers(b._bundle, controllers)
	b.addControllers(b._bundle, reverseControllers)

	// delete unmatched controllers
	b.deleteControllers(b._bundle, controllers)
}

func (b *Bag) OnlyActions(actions ...string) {
	b.checkActions(b._bundle, b._controller, actions)

	// add matched actions
	b.addActions(b._bundle, b._controller, actions)

	// delete unmatched actions
	reverseActions := b.getReverseActions(b._bundle, b._controller, actions)
	b.deleteActions(b._bundle, b._controller, reverseActions)
}

func (b *Bag) ExceptActions(actions ...string) {
	b.checkActions(b._bundle, b._controller, actions)

	// add matched actions
	reverseActions := b.getReverseActions(b._bundle, b._controller, actions)
	b.addActions(b._bundle, b._controller, reverseActions)

	// delete unmatched actions
	b.deleteActions(b._bundle, b._controller, actions)
}

// getReverseBundles 获取不包含 bundles 的所以 bundle 集合
func (b *Bag) getReverseBundles(bundles []string) []string {
	// the matched bundles
	var _bundles = make(map[string]bool, len(b.actions))

	// get all bundles
	for _bundle, _ := range b.actions {
		_bundles[_bundle] = true
	}
	// delete unmatched bundles
	for _, exBundle := range bundles {
		delete(_bundles, exBundle)
	}

	// add all matched bundles
	bundles = make([]string, len(_bundles))
	index := 0
	for _bundle, _ := range _bundles {
		bundles[index] = _bundle
		index++
	}
	return bundles
}

// getReverseActions 获取当前 bundle 中未被 controllers 包含的 controller 集合
func (b *Bag) getReverseControllers(bundle string, controllers []string) []string {
	// the matched controllers
	var _controllers = make(map[string]bool, len(b.actions[bundle]))

	// get all current bundle controllers
	for c, _ := range b.actions[bundle] {
		_controllers[c] = true
	}
	// delete unmatched controllers
	for _, exController := range controllers {
		delete(_controllers, exController)
	}

	// add all matched controllers
	controllers = make([]string, len(_controllers))
	index := 0
	for c, _ := range _controllers {
		controllers[index] = c
		index++
	}
	return controllers
}

// getReverseActions 获取当前 bundle，当前 controller 中未被 actions 包含的 action 集合
func (b *Bag) getReverseActions(bundle, controller string, actions []string) []string {
	// the matched actions
	var _actions = make(map[string]bool, len(b.actions[bundle][controller]))

	// get all current actions
	for a, _ := range b.actions[bundle][controller] {
		_actions[a] = true
	}

	// delete unmatched actions
	for _, exAction := range actions {
		delete(_actions, exAction)
	}

	// add all matched actions
	actions = make([]string, len(_actions))
	index := 0
	for c, _ := range _actions {
		actions[index] = c
		index++
	}
	return actions
}

func mapStrToSlice(m map[string]bool) []string {
	s := make([]string, len(m))
	index := 0
	for str, _ := range m {
		s[index] = str
		index++
	}
	return s
}
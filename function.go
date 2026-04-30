package main

type functionValue struct {
	declaration   *functionStmt
	closure       *environment
	isInitializer bool
}

func (f *functionValue) arity() int {
	return len(f.declaration.params)
}

func (f *functionValue) call(i *interpreter, arguments []any) (any, error) {
	env := newEnclosedEnvironment(f.closure)
	for idx, param := range f.declaration.params {
		env.define(param.lexeme, arguments[idx])

	}
	err := i.executeBlock(f.declaration.body, env)
	if ret, ok := err.(*returnValue); ok {
		if f.isInitializer {
			return f.closure.getAt(0, "this"), nil
		}
		return ret.value, nil
	} else if err != nil {
		return nil, err
	}

	if f.isInitializer {
		return f.closure.getAt(0, "this"), nil
	}

	return nil, nil
}

func (f *functionValue) bind(instance *instanceValue) *functionValue {
	environment := newEnclosedEnvironment(f.closure)
	environment.define("this", instance)

	return &functionValue{
		declaration:   f.declaration,
		closure:       environment,
		isInitializer: f.isInitializer,
	}
}

type returnValue struct {
	value any
}

func (r returnValue) Error() string {
	return "return"
}

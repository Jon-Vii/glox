package main

type functionValue struct {
	declaration *functionStmt
	closure     *environment
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
		return ret.value, nil
	}

	if err != nil {
		return nil, err
	}

	return nil, nil
}

type returnValue struct {
	value any
}

func (r returnValue) Error() string {
	return "return"
}

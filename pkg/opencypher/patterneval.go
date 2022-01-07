package opencypher

type ErrPropertiesExpected struct {
	Parameter string
}

func (e ErrPropertiesExpected) Error() string {
	return "Properties expected from parameter " + e.Parameter
}

type ExecutableNodePattern struct {
	VarName    string
	ScanLabels map[string]struct{}
	Properties map[string]Expression
}

type ExecutableEdgePattern struct {
	RtoL       bool
	VarName    string
	RelTypes   map[string]struct{}
	Properties map[string]Expression
	FromRange  *int
	ToRange    *int
}

func getExecutableProperties(ctx *EvalContext, properties *Properties) (map[string]Expression, error) {
	if properties != nil {
		if properties.Param != nil {
			v, err := ctx.GetParameter(string(*properties.Param))
			if err != nil {
				return nil, err
			}
			p, ok := v.Value.(map[string]Value)
			if !ok {
				return nil, ErrPropertiesExpected{Parameter: string(*properties.Param)}
			}
			ret := make(map[string]Expression)
			for k, v := range p {
				ret[k] = v
			}
			return ret, nil
		}
		if properties.Map != nil {
			v, err := properties.Map.Evaluate(ctx)
			if err != nil {
				return nil, err
			}
			p, _ := v.Value.(map[string]Value)
			ret := make(map[string]Expression)
			for k, v := range p {
				ret[k] = v
			}
			return ret, nil
		}
	}
	return nil, nil
}

// GetExecutablePattern processes the nodepattern and returns an
// executable node pattern that contains the resolved labels and
// properties
func (pat NodePattern) GetExecutablePattern(ctx *EvalContext) (ExecutableNodePattern, error) {
	ret := ExecutableNodePattern{
		ScanLabels: make(map[string]struct{}),
	}
	if pat.Var != nil {
		ret.VarName = string(*pat.Var)
	}
	if pat.Labels != nil {
		ret.ScanLabels = make(map[string]struct{})
		for _, label := range *pat.Labels {
			n := label.SymbolicName
			if n != nil {
				ret.ScanLabels[string(*n)] = struct{}{}
			}
		}
	}
	var err error
	ret.Properties, err = getExecutableProperties(ctx, pat.Properties)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func (pat NodePattern) Execute(ctx *EvalContext) error {
	nodeList := ctx.GetActiveNodeList()
	newNodeList, err := nodeList.ScanNodes(pat.ScanLabels, pat.Properties)
	if err != nil {
		return err
	}
	ctx.SetActiveNodeList(newNodeList)
	return nil
}

func (pat RelationshipPattern) GetExecutablePattern(ctx *EvalContext) (ExecutableEdgePattern, error) {
	ret := ExecutableEdgePattern{
		RtoL:     pat.Backwards,
		RelTypes: make(map[string]struct{}),
	}
	var err error

	if pat.Rel.Var != nil {
		ret.VarName = string(*pat.Rel.Var)
	}
	if pat.Rel.RelTypes != nil {
		ret.RelTypes = make(map[string]struct{})
		for _, label := range (*pat.Rel.RelTypes).Rel {
			n := label.SymbolicName
			if n != nil {
				ret.RelTypes[string(*n)] = struct{}{}
			}
		}
	}
	if pat.Rel.Range != nil {
		ret.FromRange, ret.ToRange, err = pat.Rel.Range.Evaluate(ctx)
		if err != nil {
			return ret, err
		}
	}

	ret.Properties, err = getExecutableProperties(ctx, pat.Rel.Properties)
	if err != nil {
		return ret, err
	}
	return ret, nil
}

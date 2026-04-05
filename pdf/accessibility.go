package pdf

import (
	"fmt"
	"io"
	"sort"
)

type AccessibilityOptions struct {
	ActualText string
	ID         string
}

type accessibilityState struct {
	dw                *DocWriter
	structTreeRoot    *structTreeRoot
	parentTree        *numberTree
	stack             []*structElem
	elementsByID      map[string]*structElem
	nextParentTreeKey int
	pageStates        map[*page]*pageAccessibilityState
}

type pageAccessibilityState struct {
	parentTreeKey int
	parents       array
	nextMCID      int
}

type numberTree struct {
	dictionaryObject
	nums map[int]writer
}

type structTreeRoot struct {
	dictionaryObject
	kids              []*structElem
	parentTree        *numberTree
	parentTreeNextKey int
}

type structElem struct {
	dictionaryObject
	s          string
	actualText string
	kids       []writer
}

type markedContentRef struct {
	page *page
	mcid int
}

type objectRef struct {
	page *page
	obj  seqGen
}

func newAccessibilityState(dw *DocWriter) *accessibilityState {
	parentTree := newNumberTree(dw.nextSeq(), 0)
	root := newStructTreeRoot(dw.nextSeq(), 0, parentTree)
	dw.file.body.add(parentTree, root)
	dw.catalog.setMarkInfo(true)
	dw.catalog.setStructTreeRoot(root)
	return &accessibilityState{
		dw:             dw,
		structTreeRoot: root,
		parentTree:     parentTree,
		elementsByID:   make(map[string]*structElem),
		pageStates:     make(map[*page]*pageAccessibilityState),
	}
}

func (s *accessibilityState) currentStructElem() *structElem {
	if len(s.stack) == 0 {
		return nil
	}
	return s.stack[len(s.stack)-1]
}

func (s *accessibilityState) push(tag string, opts AccessibilityOptions) *structElem {
	elem := s.resolveElement(s.currentStructElem(), tag, opts)
	if elem == nil {
		return nil
	}
	s.stack = append(s.stack, elem)
	return elem
}

func (s *accessibilityState) pop(elem *structElem) {
	if elem == nil || len(s.stack) == 0 {
		return
	}
	s.stack = s.stack[:len(s.stack)-1]
}

func (s *accessibilityState) resolveElement(parent *structElem, tag string, opts AccessibilityOptions) *structElem {
	if tag == "" {
		return nil
	}
	if opts.ID != "" {
		if existing := s.elementsByID[opts.ID]; existing != nil {
			if existing.actualText == "" && opts.ActualText != "" {
				existing.actualText = opts.ActualText
			}
			return existing
		}
	}

	var parentRef seqGen = s.structTreeRoot
	if parent != nil {
		parentRef = parent
	}
	elem := newStructElem(s.dw.nextSeq(), 0, tag, parentRef)
	elem.actualText = opts.ActualText
	s.dw.file.body.add(elem)
	if parent != nil {
		parent.appendKid(&indirectObjectRef{elem})
	} else {
		s.structTreeRoot.appendKid(elem)
	}
	if opts.ID != "" {
		s.elementsByID[opts.ID] = elem
	}
	return elem
}

func (s *accessibilityState) pageState(p *page) *pageAccessibilityState {
	if p == nil {
		return nil
	}
	if state := s.pageStates[p]; state != nil {
		return state
	}
	state := &pageAccessibilityState{parentTreeKey: s.nextParentTreeKey}
	s.nextParentTreeKey++
	p.setStructParents(state.parentTreeKey)
	s.pageStates[p] = state
	s.structTreeRoot.parentTreeNextKey = s.nextParentTreeKey
	return state
}

func (s *accessibilityState) associateMarkedContent(p *page, elem *structElem) int {
	if p == nil || elem == nil {
		return -1
	}
	state := s.pageState(p)
	mcid := state.nextMCID
	state.nextMCID++
	for len(state.parents) < mcid {
		state.parents = append(state.parents, null{})
	}
	state.parents = append(state.parents, &indirectObjectRef{elem})
	s.parentTree.nums[state.parentTreeKey] = state.parents
	elem.appendKid(&markedContentRef{page: p, mcid: mcid})
	return mcid
}

func (s *accessibilityState) associateObject(p *page, obj seqGen, elem *structElem) {
	if p == nil || obj == nil || elem == nil {
		return
	}
	key := s.nextParentTreeKey
	s.nextParentTreeKey++
	s.structTreeRoot.parentTreeNextKey = s.nextParentTreeKey
	if holder, ok := obj.(interface{ setStructParent(int) }); ok {
		holder.setStructParent(key)
	}
	s.parentTree.nums[key] = &indirectObjectRef{elem}
	elem.appendKid(&objectRef{page: p, obj: obj})
}

func newNumberTree(seq, gen int) *numberTree {
	tree := new(numberTree)
	tree.dictionaryObject.init(seq, gen)
	tree.nums = make(map[int]writer)
	return tree
}

func (t *numberTree) write(w io.Writer) {
	t.writeHeader(w)
	keys := make([]int, 0, len(t.nums))
	for key := range t.nums {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	nums := make(array, 0, len(keys)*2)
	for _, key := range keys {
		nums = append(nums, integer(key), t.nums[key])
	}
	t.dict["Nums"] = nums
	t.dict.write(w)
	t.writeFooter(w)
}

func newStructTreeRoot(seq, gen int, parentTree *numberTree) *structTreeRoot {
	root := new(structTreeRoot)
	root.dictionaryObject.init(seq, gen)
	root.parentTree = parentTree
	root.dict["Type"] = name("StructTreeRoot")
	root.dict["ParentTree"] = &indirectObjectRef{parentTree}
	return root
}

func (r *structTreeRoot) appendKid(elem *structElem) {
	if elem != nil {
		r.kids = append(r.kids, elem)
	}
}

func (r *structTreeRoot) write(w io.Writer) {
	r.writeHeader(w)
	switch len(r.kids) {
	case 0:
		r.dict["K"] = array{}
	case 1:
		r.dict["K"] = &indirectObjectRef{r.kids[0]}
	default:
		kids := make(array, 0, len(r.kids))
		for _, kid := range r.kids {
			kids = append(kids, &indirectObjectRef{kid})
		}
		r.dict["K"] = kids
	}
	r.dict["ParentTreeNextKey"] = integer(r.parentTreeNextKey)
	r.dict.write(w)
	r.writeFooter(w)
}

func newStructElem(seq, gen int, tag string, parent seqGen) *structElem {
	elem := new(structElem)
	elem.dictionaryObject.init(seq, gen)
	elem.s = tag
	elem.dict["Type"] = name("StructElem")
	elem.dict["S"] = name(tag)
	elem.dict["P"] = &indirectObjectRef{parent}
	return elem
}

func (e *structElem) appendKid(kid writer) {
	if kid != nil {
		e.kids = append(e.kids, kid)
	}
}

func (e *structElem) write(w io.Writer) {
	e.writeHeader(w)
	if e.actualText != "" {
		e.dict["ActualText"] = textString(e.actualText)
	}
	switch len(e.kids) {
	case 0:
		delete(e.dict, "K")
	case 1:
		e.dict["K"] = e.kids[0]
	default:
		e.dict["K"] = array(e.kids)
	}
	e.dict.write(w)
	e.writeFooter(w)
}

func (m *markedContentRef) write(w io.Writer) {
	fmt.Fprintf(w, "<<\n")
	name("Type").write(w)
	name("MCR").write(w)
	fmt.Fprintf(w, "\n")
	name("Pg").write(w)
	(&indirectObjectRef{m.page}).write(w)
	fmt.Fprintf(w, "\n")
	name("MCID").write(w)
	integer(m.mcid).write(w)
	fmt.Fprintf(w, "\n>>\n")
}

func (r *objectRef) write(w io.Writer) {
	fmt.Fprintf(w, "<<\n")
	name("Type").write(w)
	name("OBJR").write(w)
	fmt.Fprintf(w, "\n")
	name("Pg").write(w)
	(&indirectObjectRef{r.page}).write(w)
	fmt.Fprintf(w, "\n")
	name("Obj").write(w)
	(&indirectObjectRef{r.obj}).write(w)
	fmt.Fprintf(w, "\n>>\n")
}

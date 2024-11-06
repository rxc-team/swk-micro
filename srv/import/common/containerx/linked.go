package containerx

// Node 单向链表节点
type Node struct {
	Data interface{}
	Next *Node
}

// List 单向链表
type List struct {
	headNode *Node
}

// IsEmpty 判断链表是否为空
func (l *List) IsEmpty() bool {
	return l.headNode == nil
}

// Length 获取链表的长度
func (l *List) Length() int {
	currentNode := l.headNode
	count := 0

	for currentNode != nil {
		count++
		currentNode = currentNode.Next
	}

	return count
}

// Add 向链表头部添加数据
func (l *List) Add(data interface{}) *Node {
	node := &Node{Data: data}
	if l.IsEmpty() {
		l.headNode = node
		return node
	}
	node.Next = l.headNode
	l.headNode = node
	return node
}

// Append 向链表尾部添加数据
func (l *List) Append(data interface{}) *Node {
	node := &Node{Data: data}
	if l.IsEmpty() {
		l.headNode = node
		return l.headNode
	}

	currentNode := l.headNode
	for currentNode.Next != nil {
		currentNode = currentNode.Next
	}
	currentNode.Next = node

	return currentNode
}

// Insert 向链表尾部添加数据
func (l *List) Insert(i int, data interface{}) {

	if i < 0 {
		l.Add(data)
		return
	}

	if i > l.Length() {
		l.Append(data)
		return
	}

	preNode := l.headNode
	count := 0
	for count < (i - 1) {
		preNode = preNode.Next
		count++
	}

	node := &Node{Data: data}
	node.Next = preNode.Next
	preNode.Next = node
}

// Remove 删除链表中的某一个数据
func (l *List) Remove(data interface{}) {
	preNode := l.headNode
	if preNode.Data == data {
		l.headNode = preNode.Next
	} else {
		for preNode.Next != nil {
			if preNode.Next.Data == data {
				preNode.Next = preNode.Next.Next
			} else {
				preNode = preNode.Next
			}
		}
	}
}

// RemoveAtIndex 删除链表指定位置的数据
func (l *List) RemoveAtIndex(i int) {
	preNode := l.headNode
	if i <= 0 {
		l.headNode = preNode.Next
		return
	}

	if i > l.Length() {
		return
	}

	count := 0
	for count != (i-1) && preNode.Next != nil {
		count++
		preNode = preNode.Next
	}
	preNode.Next = preNode.Next.Next
}

// Contain 查看链表是否包含某一个数据
func (l *List) Contain(data interface{}) bool {
	currentNode := l.headNode
	for currentNode != nil {
		if currentNode.Data == data {
			return true
		}
		currentNode = currentNode.Next
	}
	return false
}

// ToList 返回一个数组
func (l *List) ToList() (list []interface{}) {
	if l.IsEmpty() {
		return
	}

	currentNode := l.headNode
	for currentNode != nil {
		list = append(list, currentNode.Data)
		currentNode = currentNode.Next
	}
	return
}

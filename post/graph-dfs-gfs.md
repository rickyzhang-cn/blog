<!---
title:: 图的DFS和GFS
date:: 2016-03-02 21:08
categories:: 算法
tags:: algorithm, graph 
-->

对于图G=(V, E)来说，有两种标准的表示方法，邻接链表和邻接矩阵。对于DFS和BFS，我们选择图的邻接链表这种表示。
<h4>BFS</h4>
对于BFS，我们在每个图结点u中添加额外的一些属性，u.visted表示是否已经访问，u.π表示广度优先树中的前驱结点，u.d表示记录BFS计算出的从源节点s到结点u之间的距离。
````
BFS(G, s)
	for each vertex u ∈ G.V-{s}
		u.visited=false
		u.d=∞
		u.π =NIL
	s.d=0
	s.π =NIL
	Q=∅
	ENQUEUE(Q, s)
	s.visited=true
	while Q≠∅
		u=DEQUEUE(Q)
		for each v ∈ G.Adj[u]
			if v.visited=false
				v.d=u.d+1
				v.π =u
				ENQUEUE(Q,v)
				v.visited=true
````
BFS可以运行在有向或者无向图上。在CLRS的CH22中，对于BFS产生的最短路径距离和最短路径是针对无权重的图来说的，要清楚这一点。

对于CLRS中CH22中广度优先搜索的证明，不是很难，但是图中相关证明大多都是这个思路，可以好好琢磨一下。
<h4>DFS</h4>
CLRS中关于DFS的讲述更加全面一点，涉及到了括号化定理和边的分类，这里只大概知道有这些东西就行。

这里选择纯粹点的DFS，仅仅只有DFS。
````
DFS(G)
	for each vertex u ∈ G.V
		u.π=NIL
		u.visited=false
	
	for each vertex u ∈ G.V
		if u.visited == false
			DFS-VISIT(G,u)

DFS-VISIT(G, s)
	s.visited=true
	for each v ∈ G.Adj[u]
		if v.visited == false
			DFS(G,v)
			v.π=s
````
不论有向还是无向图，均可以使用DFS。DFS采用递归来实现，非递归采用FILO的stack也很好实现。
<h4>Topological Sort</h4>
对于一个有向无环图(也就是DAG) G=(V, E)来说，其拓扑排序是G中所有结点的一种线性次序，该次序满足如下条件：如果图G包含边(u, v)，则结点u在拓扑排序中处于结点v的前面。

对于拓扑排序，我们采用修正后的DFS，只需要在DFS中添加一个stack即可。在对一个节点的每个unvisited的邻接节点进行DFS后，将其push到stack中即可保证如果图G包含边(u, v)，则结点u在拓扑排序中处于结点v的前面。
````
TOPOLOGICAL-SORT(G)
	for each vertex u ∈ G.V
		u.visited=false
	STACK=∅
	for each vertex u ∈ G.V
		if u.visited == false
			DFS-VISIT(G,u,STACK)
	
	POP STACK until empty


DFS-VISIT(G, u, STACK)
	u.visited=true
	for each v ∈ G.Adj[u]
		if v.visited == false
			DFS(G,v)
        push u into STACK
````
DAG的拓扑排序可以是某些简单DP问题的基础模型。

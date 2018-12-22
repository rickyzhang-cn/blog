<!---
title:: Dijkstra算法
date:: 2016-03-08 21:08
categories:: 算法
tags:: algorithm, graph 
-->

Dijkstra算法解决的是带权重的有向图上单源最短路径问题，该算法要求所有边的权重都为非负值。

Dijkstra算法解决的是单源最短路径问题，Bellman-Ford算法也可以解决，而且Bellman-Ford算法可以运行在有权重和为负值的环的图上，Dijkstra具有更快的运行效率。对于DAG，这种自带约束条件，我们有更简单的方法来解决这个问题，基于拓扑排序来解决。

对于单源最短路径，最基本的操作是relaxation操作。代码很简单：
```
INITIALIZE-SINGLE-SOURCE(G, s)
	for each vertex v ∈ G.V
		v.d=∞
		v.π=NIL
	s.d=0
	
RELAX(u,v,w)
	if v.d > u.d+w(u,v)
		v.d=u.d+w(u,v)
		v.π=u
```
属性v.d用来记录从源节点s到结点v的最短路径权重的上界。

最短路径和松弛操作的性质提供了单源最短路径算法的证明。

Dijkstra算法代码：
```
DIJKSTRA(G,w,s)
	INITIALIZE_SINGLE-SOURCE(G,s)
	S=∅
	Q=G.V
	while Q≠∅
		u=EXTRACT-MIN(Q)
		S=S∪{u}
		for each vertex v∈G.Adj[u]
			RELAX(u,v,w)
```
Dijkstra算法的代码执行过程很简单，但是很巧妙。最重要的就是要理解下面的循环不等式：

在算法第4-8行的while语句的每次循环开始前，对于每个结点v∈S，有v.d=δ(s,v)。

关于循环不等式的证明，结合算法中的最小优先队列和松弛操作的性质，通过反证法可以证明，具体过程参见CLRS的相关章节。

Dijkstra算法的运行时间和最小优先队列的选择有关，从历史的角度来看，斐波那契堆的提出动机就是因为人们观察到Dijkstra算法调用DECREASE-KEY操作通常比EXTRACT-MIN操作更多。

Dijkstra算法的流程其实比较简单，最重要的是Dijkstra算法正确性的证明！
<!---
title:: 最小生成树
date:: 2016-03-04 21:08
categories:: 算法
tags:: algorithm, graph 
-->

MST运行在连通无向图上，每条边带有权重。

Kruskal和Prim算法都属于贪心算法，贪心策略可以由下面的通用方法来表述。该通用方法在每个时刻生长最小生成树的一条边，并在整个策略的实施过程中，管理一个遵守下述循环不变式的边集合A：在每遍循环前，A是某棵最小生成树的一个子集。

无向图G=(V, E)的一个切割(S, V-S)是集合V的一个划分。如果一条边(u, v)∈E的一个端点位于集合S，另一个端点位于集合V-S，则称该条边横跨切割(S, V-S)。如果集合A中不存在横跨该切割的边，则称该切割尊重集合A。在横跨一个切割的所有的边中，权重最小的边称为轻量级边。
<h4>Kruskal算法</h4>
在Kruskal算法中，集合A是一个森林，其结点就是给定图的结点。每次加入到集合A中的安全边永远是权重最小的连接两个不同分量的边。

Kruskal算法的正确性可以使用CLRS中CH23的推论23.2证明。

Kruskal算法文字描述很简单，代码的话，Kruskal算法使用UNION过程来对两棵树进行合并。
<h4>Prim算法</h4>
在Prim算法中，集合A则是一颗树，每次加入到A中的安全边永远是连接A和A之外某个结点的边中权重最小的边。
<pre class="brush:plain">MST-PRIM(G,w,r)
	for each u∈G.V
		u.key=∞
		u.π=NIL
	r.key=0
	Q=G.V
	while Q≠∅
		u=EXTRACT-MIN(Q)
		for each v∈G.Adj[u]
		if v and w(u,v)&lt;v.key
			v.π=u
			v.key=w(u,v)</pre>
算法中的while的循环不等式为：

1.A={(v, v.π): v∈V-{r}-Q}

2.已经加入到最小生成树的结点为集合V-Q

3.对于所有的结点v∈Q，如果v.π≠NIL，则v.key&lt;∞并且v.key是连接点v和最小生成树中某个结点的轻量级边(v, v.π)的权重

证明了循环不等式结合CLRS中的推论23.2可以证明Prim算法的正确性。

Prim算法的运行时间和最小优先队列的实现有关，可以使用二叉最小优先队列，斐波那契堆等等。
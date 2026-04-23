// ====== 1. 模态框自动打开逻辑 ======
document.addEventListener('DOMContentLoaded', function() {
    // 自动打开项目编辑对话框
    if (window.AppData && window.AppData.editMode) {
        const editModal = document.getElementById('edit-modal');
        if (editModal && !editModal.hasAttribute('open')) {
            editModal.showModal();
        }
    }

    // 自动打开版本编辑对话框
    if (window.AppData && window.AppData.editVersionMode) {
        const editVersionModal = document.getElementById('edit-version-modal');
        if (editVersionModal) {
            editVersionModal.showModal();
        }
    }
});

// ====== 2. 子项目分类自动同步功能 ======
document.addEventListener('DOMContentLoaded', function() {
    // 处理函数：根据选择的父项目锁定或解锁分类选择
    function syncCategory(parentSelectId, categorySelectId) {
        const parentSelect = document.getElementById(parentSelectId);
        const categorySelect = document.getElementById(categorySelectId);
        let categoryLocked = false;

        if (parentSelect && categorySelect) {
            parentSelect.addEventListener('change', function() {
                const selectedOption = this.options[this.selectedIndex];
                const parentCategory = selectedOption.getAttribute('data-category');

                if (parentCategory && parentCategory !== '') {
                    // 选择了有效父项目：强制同步分类并禁用手动修改
                    categorySelect.value = parentCategory;
                    categoryLocked = true;
                    categorySelect.style.backgroundColor = '#f3f4f6';
                    categorySelect.style.cursor = 'not-allowed';
                    categorySelect.title = '子项目分类必须与父项目一致';
                } else {
                    // 顶级项目：允许自由选择分类
                    categoryLocked = false;
                    categorySelect.style.backgroundColor = '';
                    categorySelect.style.cursor = '';
                    categorySelect.title = '';
                }
            });

            // 拦截点击和键盘事件，防止在锁定状态下修改
            const preventChange = function(e) {
                if (categoryLocked) e.preventDefault();
            };
            categorySelect.addEventListener('mousedown', preventChange);
            categorySelect.addEventListener('keydown', preventChange);

            // 初始化触发一次
            if (parentSelect.value !== '0' && parentSelect.value !== '') {
                parentSelect.dispatchEvent(new Event('change'));
            }
        }
    }

    // 应用到新增和编辑表单
    syncCategory('add-parent-id', 'add-category');
    syncCategory('edit-parent-id', 'edit-category');
});

// ====== 3. 子项目树形结构功能 ======
(function() {
    // 检查 AppData 是否可用
    if (!window.AppData || !window.AppData.projects) return;

    const projects = window.AppData.projects;
    const currentProjectId = window.AppData.currentProjectId || 0;
    const TREE_STORAGE_KEY = 'tree_nodes_state';

    // 状态管理：保存/恢复树的展开收起状态
    function saveTreeState() {
        const state = {};
        document.querySelectorAll('.tree-children').forEach(childrenDiv => {
            const nodeId = childrenDiv.getAttribute('data-node-id');
            if (nodeId) {
                state[nodeId] = childrenDiv.style.display === 'block';
            }
        });
        sessionStorage.setItem(TREE_STORAGE_KEY, JSON.stringify(state));
    }

    function restoreTreeState() {
        const saved = sessionStorage.getItem(TREE_STORAGE_KEY);
        if (!saved) return;
        const state = JSON.parse(saved);
        document.querySelectorAll('.tree-children').forEach(childrenDiv => {
            const nodeId = childrenDiv.getAttribute('data-node-id');
            if (nodeId && state.hasOwnProperty(nodeId)) {
                const toggle = childrenDiv.previousElementSibling?.querySelector('.tree-toggle');
                if (toggle) {
                    const isOpen = state[nodeId];
                    childrenDiv.style.display = isOpen ? 'block' : 'none';
                    toggle.innerHTML = isOpen ? '▼' : '▶';
                }
            }
        });
    }

    // 构建递归对象
    function buildTree(items, parentId = 0) {
        return items
            .filter(item => item.parentId === parentId)
            .map(item => ({
                ...item,
                children: buildTree(items, item.id)
            }));
    }

    // DOM 渲染引擎
    function renderTree(nodes, container, level = 0) {
        nodes.forEach(node => {
            const hasChildren = node.children && node.children.length > 0;
            const isCurrent = node.id === currentProjectId;

            const itemDiv = document.createElement('div');
            itemDiv.className = 'tree-item';

            const contentDiv = document.createElement('div');
            contentDiv.className = 'tree-item-content';

            // 展开图标
            const toggle = document.createElement('span');
            toggle.className = 'tree-toggle';
            toggle.innerHTML = hasChildren ? '▼' : ''; 
            contentDiv.appendChild(toggle);

            // 标题链接
            const link = document.createElement('a');
            link.href = '/?id=' + node.id;
            link.textContent = node.name;
            if (isCurrent) link.classList.add('active');
            link.style.flex = '1';
            contentDiv.appendChild(link);

            // 添加子项目快捷键
            const addChildBtn = document.createElement('span');
            addChildBtn.className = 'add-child-btn';
            addChildBtn.textContent = '+ 子项目';
            addChildBtn.onclick = (e) => {
                e.stopPropagation();
                openAddChildModal(node.id);
            };
            contentDiv.appendChild(addChildBtn);

            itemDiv.appendChild(contentDiv);
            container.appendChild(itemDiv);

            if (hasChildren) {
                const childrenDiv = document.createElement('div');
                childrenDiv.className = 'tree-children';
                childrenDiv.setAttribute('data-node-id', node.id);
                childrenDiv.style.display = 'block'; 
                renderTree(node.children, childrenDiv, level + 1);
                container.appendChild(childrenDiv);

                toggle.onclick = () => {
                    const isHidden = childrenDiv.style.display === 'none';
                    childrenDiv.style.display = isHidden ? 'block' : 'none';
                    toggle.innerHTML = isHidden ? '▼' : '▶';
                    saveTreeState();
                };
            }
        });
    }

    // 暴露给全局
    window.openAddChildModal = function(parentId) {
        const modal = document.getElementById('add-modal');
        const parentSelect = modal.querySelector('select[name="parent_id"]');
        if (parentSelect) {
            parentSelect.value = parentId;
            parentSelect.dispatchEvent(new Event('change'));
        }
        modal.showModal();
    };

    document.addEventListener('DOMContentLoaded', function() {
        const mapping = [
            { cat: '科研', container: 'tree-research' },
            { cat: '个人项目', container: 'tree-personal' },
            { cat: '其他', container: 'tree-other' }
        ];

        mapping.forEach(m => {
            const catProjects = projects.filter(p => p.category === m.cat);
            const treeData = buildTree(catProjects);
            const container = document.getElementById(m.container);
            if (container && treeData.length > 0) {
                renderTree(treeData, container);
            }
        });

        restoreTreeState();
    });
})();

// ====== 4. 菜单栏分类展开/折叠记忆 ======
document.addEventListener('DOMContentLoaded', function() {
    const STORAGE_KEY = 'sidebar_details_state';

    function saveState() {
        const state = {};
        document.querySelectorAll('.sidebar-details').forEach((details, index) => {
            state[index] = details.hasAttribute('open');
        });
        sessionStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    }

    function restoreState() {
        const saved = sessionStorage.getItem(STORAGE_KEY);
        if (!saved) return;
        const state = JSON.parse(saved);
        document.querySelectorAll('.sidebar-details').forEach((details, index) => {
            if (state[index]) details.setAttribute('open', '');
            else details.removeAttribute('open');
        });
    }

    restoreState();
    document.querySelectorAll('.sidebar-details').forEach(details => {
        details.addEventListener('toggle', saveState);
    });
});

// ====== 5. 侧边栏宽度调整功能 ======
document.addEventListener('DOMContentLoaded', function() {
    const SIDEBAR_WIDTH_KEY = 'sidebar_width';
    const sidebar = document.getElementById('sidebar');
    const resizer = document.getElementById('sidebarResizer');
    
    if (!sidebar || !resizer) return;

    let isResizing = false;
    let startX = 0;
    let startWidth = 0;

    const savedWidth = localStorage.getItem(SIDEBAR_WIDTH_KEY);
    if (savedWidth) {
        sidebar.style.width = savedWidth + 'px';
    }

    resizer.addEventListener('mousedown', function(e) {
        isResizing = true;
        startX = e.clientX;
        startWidth = sidebar.offsetWidth;
        document.body.style.cursor = 'col-resize';
        
        document.addEventListener('mousemove', onMouseMove);
        document.addEventListener('mouseup', onMouseUp);
        
        e.preventDefault();
    });

    function onMouseMove(e) {
        if (!isResizing) return;
        const newWidth = Math.max(200, Math.min(600, startWidth + (e.clientX - startX)));
        sidebar.style.width = newWidth + 'px';
    }

    function onMouseUp() {
        if (!isResizing) return;
        isResizing = false;
        document.body.style.cursor = '';
        localStorage.setItem(SIDEBAR_WIDTH_KEY, sidebar.offsetWidth);
        document.removeEventListener('mousemove', onMouseMove);
        document.removeEventListener('mouseup', onMouseUp);
    }
});

// ====== 6. 标签页切换功能 ======
document.addEventListener('DOMContentLoaded', function() {
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');
    
    tabBtns.forEach(btn => {
        btn.addEventListener('click', function() {
            const tabName = this.getAttribute('data-tab');
            
            // 移除所有按钮的 active 状态
            tabBtns.forEach(b => b.classList.remove('active'));
            // 移除所有内容区域的 active 状态
            tabContents.forEach(content => content.classList.remove('active'));
            
            // 给当前按钮添加 active 状态
            this.classList.add('active');
            // 给对应的内容区域添加 active 状态
            const activeContent = document.getElementById('tab-' + tabName);
            if (activeContent) {
                activeContent.classList.add('active');
            }
        });
    });
});
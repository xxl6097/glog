#!/bin/bash
#version=$(if [ "$(git describe --tags --abbrev=0 2>/dev/null)" != "" ]; then git describe --tags --abbrev=0; else git log --pretty=format:'%h' -n 1; fi)
#version=$(git tag -l "v*" --sort=-creatordate | head -n 1)
version=$(git tag -l "[0-99]*.[0-99]*.[0-99]*" --sort=-creatordate | head -n 1)
#git tag --sort=-creatordate | head -n 1
#git tag -l "v*" --sort=-creatordate | head -n 1
#git tag -l "v[0-99][0-99].[0-99][0-99].[0-99][0-99]" --sort=-v:refname | head -n 1
#git tag -l "v*.*.*" --sort=-v:refname | head -n 1
# git tag -l "[0-99]*.[0-99]*.[0-99]*" --sort=-creatordate | head -n 1
function upgradeVersion() {
  if [ "$version" = "" ]; then
    version="0.0.0"
  else
    v3=$(echo $version | awk -F'.' '{print($3);}')
    v2=$(echo $version | awk -F'.' '{print($2);}')
    v1=$(echo $version | awk -F'.' '{print($1);}')
    if [[ $(expr $v3 \>= 99) == 1 ]]; then
      v3=0
      if [[ $(expr $v2 \>= 99) == 1 ]]; then
        v2=0
        v1=$(expr $v1 + 1)
      else
        v2=$(expr $v2 + 1)
      fi
    else
      v3=$(expr $v3 + 1)
    fi
    version="$v1.$v2.$v3"
  fi
}

function todir() {
  pwd
}

function pull() {
  todir
  echo "git pull"
  git pull
}

# shellcheck disable=SC2120
function forcepull() {
  todir
  echo "git fetch --all && git reset --hard origin/$1 && git pull"
  git fetch --all && git reset --hard origin/$1 && git pull
}

function tag() {
  echo "===>${version}"
  git add .
  git commit -m "release ${version}"
  git tag -a $version -m "release ${version}"
  git push origin $version
}

function push() {
  commit=""
  if [ ! -n "$1" ]; then
    commit="$(date '+%Y-%m-%d %H:%M:%S') by ${USER}"
  else
    commit="$1 by ${USER}"
  fi
  echo $commit
  git add .
  git commit -m "${version} $commit"
  echo "提交代码"
  git push
#  git push -u origin main
#  echo "打tag标签"
#  tag
}

function main_pre() {
  #1. 更新版本号
  upgradeVersion
}


function tagAndGitPush() {
    read -p "请输入标签名称: " commit
    commit="$commit $(date '+%Y-%m-%d %H:%M:%S') by ${USER}"
    vtag="$(date '+%Y.%m.%d.%H.%M.%S')"
    git add .
    git commit -m "${commit}"
    git tag -a v$vtag -m "${commit}"
    git push origin v$vtag
}

function forceBranch() {
    # 获取所有分支列表（包含远程分支）
    git fetch origin > /dev/null 2>&1
    # shellcheck disable=SC2207
    branches=($(git branch -r | grep -v "HEAD" | sed 's/^* //' | sed 's/remotes\///'))
    #git branch origin/test002
    # 获取所有远程分支信息
#    git fetch origin --prune > /dev/null 2>&1
#
#    # 获取所有本地和远程分支（过滤 HEAD 和重复项）
#    branches=$(git branch -a |
#        grep -v 'HEAD' |
#        sed 's/^\*\? *//;s/remotes\/origin\///' |
#        awk '!seen[$0]++' |
#        grep -vE '^origin/(main|master)$')  # 过滤远程默认分支

    # 生成分支菜单
    echo "可更新的分支列表："
    select branch in "${branches[@]}"; do
        if [[ -n "$branch" ]]; then
            if [ $1 -eq 0 ]; then
                echo "正在更新分支：$branch"
                git checkout "$branch" > /dev/null 2>&1
                git pull origin "$branch"
            else
                echo "正在更新分支（强制）：$branch"
                git checkout "$branch" > /dev/null 2>&1
                forcepull "$branch"
            fi
            break
        else
            echo "输入无效，请重新选择。"
        fi
    done
}

function forcePullCurrent() {
  forcepull "$(git branch)"
}

function pullMenu() {
    echo "1. 强制更新"
    echo "2. 普通更新"
    echo "3. 分支更新"
    echo "4. 分支更新(强制)"
    echo "请输入编号:"
    read index
    clear
    case "$index" in
    [1]) (forcePullCurrent);;
    [2]) (pull);;
    [3]) (forceBranch 0);;
    [4]) (forceBranch 1);;
    *) echo "exit" ;;
  esac
}

function createBranch() {
    read -p "请输入分支名称: " branchName
    git branch "$branchName"
    commit="$(date '+%Y-%m-%d %H:%M:%S') by ${USER}"
    git add .
    git commit -m "new branch $branchName created ${commit}"
    #-u 首次推送，-f 强制推送
    git push -u origin "$branchName"
}
function deleteBranch() {
    # 获取远程分支列表（过滤 origin/HEAD 无效指针）
    # shellcheck disable=SC2207
    remote_branches=($(git branch -r | grep -v "HEAD" | sed 's/origin\///' | awk '{print $1}'))

    # 检查是否有远程分支
    if [ ${#remote_branches[@]} -eq 0 ]; then
        echo "无远程分支可删除"
        exit 0
    fi

    # 显示分支菜单
    echo "可删除的远程分支列表："
    PS3="请输入要删除的分支编号（输入 q 退出）: "
    select branch in "${remote_branches[@]}"; do
        case $REPLY in
            q|Q)
                echo "退出操作"
                exit 0
                ;;
            *)
                if [[ "$REPLY" =~ ^[0-9]+$ ]] && [ "$REPLY" -le ${#remote_branches[@]} ]; then
                    selected_branch=${remote_branches[$REPLY-1]}
                    echo -n "确认删除远程分支 origin/$selected_branch ？(y/n): "
                    read confirm
                    if [[ $confirm =~ [Yy] ]]; then
                        git push origin --delete "$selected_branch"
                        # 检查删除结果
                        if [ $? -eq 0 ]; then
                            echo "删除成功"
                            git fetch --prune  # 清理本地缓存
                            break
                        else
                            echo "删除失败，请检查权限或网络"
                        fi
                    else
                        echo "取消删除"
                    fi
                else
                    echo "输入无效，请重新选择"
                fi
                ;;
        esac
    done
}

function switchBranch() {
    # 获取所有本地和远程分支（过滤 origin/HEAD 指针）
    # shellcheck disable=SC2207
    #branches=($(git branch -a | grep -v "HEAD" | sed 's/remotes\/origin\///' | awk '{print $1}' | sort -u))
    branches=($(git branch -r | grep -v "HEAD" | sed 's/origin\///' | awk '{print $1}'))

    # 检查是否有可用分支
    if [ ${#branches[@]} -eq 0 ]; then
        echo "无可用分支"
        exit 1
    fi

    # 显示分支菜单
    echo "可用分支列表："
    PS3="请输入要切换的分支编号（输入 q 退出）: "
    select branch in "${branches[@]}"; do
        case $REPLY in
            q|Q)
                echo "退出操作"
                exit 0
                ;;
            *)
                # 输入有效性校验
                if [[ "$REPLY" =~ ^[0-9]+$ ]] && [ "$REPLY" -le ${#branches[@]} ]; then
                    selected_branch=${branches[$REPLY-1]}

                    git checkout "$selected_branch"

                    # 检查切换结果
                    if [ $? -eq 0 ]; then
                        echo "已切换到分支：$selected_branch"
                        exit 0
                    else
                        echo "切换失败，请检查未提交的修改（可使用 git stash 暂存）"
                        exit 1
                    fi
                else
                    echo "输入无效，请重新选择"
                fi
                ;;
        esac
    done
}

function branchMenu() {
    echo "1. 创建分支"
    echo "2. 删除分支"
    echo "3. 切换分支"
    echo "请输入编号:"
    read index
    clear
    case "$index" in
    [1]) (createBranch);;
    [2]) (deleteBranch);;
    [3]) (switchBranch);;
    *) echo "exit" ;;
  esac
}

function m() {
    echo "1. 快速提交"
    echo "2. 项目更新"
    echo "3. 项目标签"
    echo "4. 分支管理"
    echo "请输入编号:"
    read index
    clear
    case "$index" in
    [1]) (push);;
    [2]) (pullMenu);;
    [3]) (tagAndGitPush);;
    [4]) (branchMenu);;
    *) echo "exit" ;;
  esac
}

function main() {
  main_pre
  m
}

function test() {
    git fetch origin > /dev/null 2>&1
    # shellcheck disable=SC2207
    branches=($(git branch -r | grep -v "HEAD" | sed 's/^* //' | sed 's/remotes\///'))
    # shellcheck disable=SC2128
    echo "$branches"# 生成分支菜单
    echo "可更新的分支列表："
    select branch in "${branches[@]}"; do
        if [[ -n "$branch" ]]; then
            if [ $1 -eq 0 ]; then
                echo "正在更新分支：$branch"
                git checkout "$branch" > /dev/null 2>&1
                git pull origin "$branch"
            else
                echo "正在更新分支（强制）：$branch"
                git checkout "$branch" > /dev/null 2>&1
                forcepull "$branch"
            fi
            break
        else
            echo "输入无效，请重新选择。"
        fi
    done
}

main
#test

